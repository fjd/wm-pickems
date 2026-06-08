package notify

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/floholz/wm-pickems/internal/push"
	"github.com/floholz/wm-pickems/internal/users"
)

// League-chat push (v2). High-frequency and real-time, so it runs on its own
// cron (every minute by default; CHAT_NOTIFY_CRON overrides) keyed off real wall
// time — not the dev sim clock. It pushes one summary per private league to
// members with unread messages, with a per-(league,user) cooldown so an active
// room can't spam.
const (
	chatEvent      = "league_chat"
	chatCooldown   = 10 * time.Minute // at most one chat push per league/user per window
	chatMaxAge     = time.Hour        // don't notify about chats quiet longer than this
	chatPreviewMax = 120
)

// registerChatCron schedules the chat-notification pass on the PB cron
// (minute-granularity; CHAT_NOTIFY_CRON overrides). No-op without push.
func registerChatCron(app core.App, r *Runner) {
	if r.push == nil || !r.push.Enabled() {
		log.Printf("[chat] push not configured — chat notifications off")
		return
	}
	expr := os.Getenv("CHAT_NOTIFY_CRON")
	if expr == "" {
		expr = "* * * * *"
	}
	app.Cron().MustAdd("chat-notify", expr, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		r.chatPass(ctx)
	})
	log.Printf("[chat] notifications enabled (%s)", expr)
}

// chatPass runs one notification sweep, returning how many pushes it sent.
func (r *Runner) chatPass(ctx context.Context) int {
	ncol, err := r.notificationsCol()
	if err != nil {
		return 0
	}
	base := r.base()
	bucket := time.Now().Unix() / int64(chatCooldown.Seconds())

	leagues, err := r.app.FindRecordsByFilter("leagues", "inviteCode != 'GLOBAL'", "", 0, 0)
	if err != nil {
		return 0
	}
	names := map[string]string{} // userId -> display name (per-pass cache)
	sent := 0

	for _, lg := range leagues {
		lid := lg.Id
		latestList, _ := r.app.FindRecordsByFilter("league_messages",
			"league = {:l} && deleted = false", "-created", 1, 0, dbx.Params{"l": lid})
		if len(latestList) == 0 {
			continue
		}
		latest := latestList[0]
		latestAt := latest.GetDateTime("created").Time()
		if time.Since(latestAt) > chatMaxAge {
			continue // chat's gone quiet — leave it to the unread badge / email digest
		}

		members, _ := r.app.FindRecordsByFilter("league_members",
			"league = {:l}", "", 0, 0, dbx.Params{"l": lid})
		leagueName := lg.GetString("name")
		senderName := cachedName(r.app, names, latest.GetString("user"))
		preview := truncateRunes(latest.GetString("text"), chatPreviewMax)

		for _, mem := range members {
			uid := mem.GetString("user")
			// Caught up (read >= latest) → skip. Posting marks the sender read, so
			// this also excludes the author and anyone actively in the chat.
			if lr := r.lastReadAt(lid, uid); !lr.IsZero() && !lr.Before(latestAt) {
				continue
			}
			u, err := r.app.FindRecordById("users", uid)
			if err != nil || users.IsBot(u) {
				continue
			}
			if !prefEnabled(u, chatEvent, "push") {
				continue
			}
			subs, err := push.Subscriptions(r.app, uid)
			if err != nil || len(subs) == 0 {
				continue
			}
			unread := r.chatUnread(lid, uid)
			if unread == 0 {
				continue
			}
			dedupKey := fmt.Sprintf("chat:%s:%s:%d", lid, uid, bucket)
			if r.alreadySent(dedupKey, "push") {
				continue
			}

			title := leagueName
			if unread > 1 {
				title = fmt.Sprintf("%s · %d new", leagueName, unread)
			}
			n := push.Notification{
				Title: title,
				Body:  senderName + ": " + preview,
				URL:   toPath(base.url + "/leagues/" + lid + "/chat"),
				Tag:   "chat:" + lid,
				Icon:  "/icons/notif/default.png",
			}
			rec := newLedgerRow(ncol, uid, chatEvent, dedupKey, "push")
			if err := r.app.Save(rec); err != nil {
				continue // unique (dedupKey,channel) — another pass already claimed it
			}
			ok, sendErr := r.sendPush(ctx, subs, n)
			if ok > 0 {
				rec.Set("status", "sent")
				rec.Set("sentAt", time.Now().UTC())
				sent++
			} else {
				rec.Set("status", "failed")
				if sendErr != nil {
					rec.Set("error", sendErr.Error())
				}
			}
			_ = r.app.Save(rec)
		}
	}
	return sent
}

func (r *Runner) lastReadAt(leagueID, userID string) time.Time {
	rec, err := r.app.FindFirstRecordByFilter("league_reads",
		"league = {:l} && user = {:u}", dbx.Params{"l": leagueID, "u": userID})
	if err != nil {
		return time.Time{}
	}
	return rec.GetDateTime("lastRead").Time()
}

// chatUnread counts a user's unread, non-deleted messages in a league (capped).
func (r *Runner) chatUnread(leagueID, userID string) int {
	filter := "league = {:l} && user != {:u} && deleted = false"
	params := dbx.Params{"l": leagueID, "u": userID}
	if rec, err := r.app.FindFirstRecordByFilter("league_reads",
		"league = {:l} && user = {:u}", dbx.Params{"l": leagueID, "u": userID}); err == nil {
		filter += " && created > {:s}"
		params["s"] = rec.GetDateTime("lastRead").String()
	}
	recs, err := r.app.FindRecordsByFilter("league_messages", filter, "", 100, 0, params)
	if err != nil {
		return 0
	}
	return len(recs)
}

func cachedName(app core.App, cache map[string]string, uid string) string {
	if n, ok := cache[uid]; ok {
		return n
	}
	n := "Someone"
	if u, err := app.FindRecordById("users", uid); err == nil {
		if nm := u.GetString("name"); nm != "" {
			n = nm
		}
	}
	cache[uid] = n
	return n
}

func truncateRunes(s string, max int) string {
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	return string(rs[:max]) + "…"
}
