package notify

import (
	"context"
	"fmt"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/floholz/wm-pickems/internal/scoring"
)

// stageOrder is the canonical tournament progression; stageName maps the stored
// codes to human labels used in emails.
var stageOrder = []string{"group", "R32", "R16", "QF", "SF", "3RD", "FINAL"}

var stageName = map[string]string{
	"group": "Group Stage",
	"R32":   "Round of 32",
	"R16":   "Round of 16",
	"QF":    "Quarter-finals",
	"SF":    "Semi-finals",
	"3RD":   "Third-place Play-off",
	"FINAL": "Final",
}

func (r *Runner) notificationsCol() (*core.Collection, error) {
	return r.app.FindCollectionByNameOrId("notifications")
}

// detectStageStarting emails everyone when a stage's first kickoff enters the
// lead window. One email per stage (dedup stage_starting:<stage>).
func (r *Runner) detectStageStarting(ctx context.Context, res *Result, now time.Time,
	lead time.Duration, matches []*core.Record, recipients []*core.Record, base baseInfo) error {

	// Earliest kickoff per stage.
	starts := map[string]time.Time{}
	for _, m := range matches {
		st := m.GetString("stage")
		ko := m.GetDateTime("kickoff").Time()
		if cur, ok := starts[st]; !ok || ko.Before(cur) {
			starts[st] = ko
		}
	}

	ncol, err := r.notificationsCol()
	if err != nil {
		return err
	}
	for _, st := range stageOrder {
		start, ok := starts[st]
		if !ok || !inLeadWindow(now, start, lead) {
			continue
		}
		data := tplData{
			StageName: stageName[st],
			StartsIn:  humanizeDur(start.Sub(now)),
			WhenText:  formatKickoff(start),
			CTAText:   "Open your tips",
			CTAUrl:    base.url + "/tips",
		}
		for _, u := range recipients {
			r.dispatch(ctx, res, ncol, u, "stage_starting", "stage_starting:"+st, data)
		}
	}
	return nil
}

// detectForecastReminder nudges users whose Forecast is incomplete as the global
// lock (tournament first kickoff) approaches.
func (r *Runner) detectForecastReminder(ctx context.Context, res *Result, now time.Time,
	lead time.Duration, matches []*core.Record, recipients []*core.Record, base baseInfo) error {

	if len(matches) == 0 {
		return nil
	}
	start := matches[0].GetDateTime("kickoff").Time() // sorted by kickoff asc
	if !inLeadWindow(now, start, lead) {
		return nil
	}
	groupCount, err := r.app.CountRecords("tournament_groups")
	if err != nil {
		return err
	}
	ncol, err := r.notificationsCol()
	if err != nil {
		return err
	}
	data := tplData{
		StartsIn: humanizeDur(start.Sub(now)),
		WhenText: formatKickoff(start),
		CTAText:  "Finish your Forecast",
		CTAUrl:   base.url + "/forecast",
	}
	for _, u := range recipients {
		if !r.forecastIncomplete(u.Id, int(groupCount)) {
			continue
		}
		r.dispatch(ctx, res, ncol, u, "forecast_reminder", "forecast_reminder:"+u.Id, data)
	}
	return nil
}

// detectTipsReminder sends a per-user digest of upcoming matches (within the
// lead window) the user hasn't tipped. Dedup is per (user, match) so each match
// is reminded at most once, while the email batches all newly-missing matches.
func (r *Runner) detectTipsReminder(ctx context.Context, res *Result, now time.Time,
	lead time.Duration, matches []*core.Record, recipients []*core.Record, base baseInfo) error {

	windowEnd := now.Add(lead)
	var upcoming []*core.Record
	for _, m := range matches {
		ko := m.GetDateTime("kickoff").Time()
		if m.GetString("status") == "scheduled" && ko.After(now) && !ko.After(windowEnd) {
			upcoming = append(upcoming, m)
		}
	}
	if len(upcoming) == 0 {
		return nil
	}
	ncol, err := r.notificationsCol()
	if err != nil {
		return err
	}
	names := r.teamNames()

	for _, u := range recipients {
		if !prefEnabled(u, "tips_reminder") {
			continue
		}
		var newMissing []*core.Record
		for _, m := range upcoming {
			if r.hasTip(u.Id, m.Id) {
				continue
			}
			if r.alreadySent("tips_reminder:" + u.Id + ":" + m.Id) {
				continue
			}
			newMissing = append(newMissing, m)
		}
		if len(newMissing) == 0 {
			continue
		}
		res.Considered++

		lines := make([]matchLine, 0, len(newMissing))
		for _, m := range newMissing {
			lines = append(lines, matchLine{
				Home:     r.teamLabel(m, "homeTeam", "homeLabel", names),
				Away:     r.teamLabel(m, "awayTeam", "awayLabel", names),
				WhenText: formatKickoff(m.GetDateTime("kickoff").Time()),
			})
		}
		data := tplData{
			AppName:     base.appName,
			SettingsUrl: base.url + "/settings",
			CTAText:     "Enter your tips",
			CTAUrl:      base.url + "/tips",
			Count:       len(lines),
			Matches:     lines,
		}
		subject, html, text, rerr := render("tips_reminder", data)
		if rerr != nil {
			res.Failed++
			continue
		}
		mid, serr := r.sender.Send(ctx, mailerMessage(u, subject, html, text))

		// One ledger row per match so each is deduped independently; they share
		// the provider message id of the single digest email.
		status, errStr := "sent", ""
		if serr != nil {
			status, errStr = "failed", serr.Error()
			res.Failed++
		} else {
			res.Sent++
		}
		for _, m := range newMissing {
			rec := core.NewRecord(ncol)
			rec.Set("user", u.Id)
			rec.Set("event", "tips_reminder")
			rec.Set("dedupKey", "tips_reminder:"+u.Id+":"+m.Id)
			rec.Set("channel", "email")
			rec.Set("status", status)
			rec.Set("providerMessageId", mid)
			rec.Set("error", errStr)
			if status == "sent" {
				rec.Set("sentAt", time.Now().UTC())
			}
			_ = r.app.Save(rec)
		}
	}
	return nil
}

// detectResultsRecap sends a once-daily digest (gated to the recap hour by the
// caller) summarising points earned from matches finalized in the last 24h plus
// the user's current league rankings.
func (r *Runner) detectResultsRecap(ctx context.Context, res *Result, now time.Time,
	matches []*core.Record, recipients []*core.Record, base baseInfo) error {

	// Recently-resolved matches = finished and kicked off within the last 24h.
	// Keying off kickoff (fixed schedule data) rather than finalizedAt keeps
	// consecutive daily windows gap-free and makes the recap behave correctly
	// under the dev virtual clock (finalizedAt is stamped with real wall-time).
	since := now.Add(-24 * time.Hour)
	finalized := map[string]bool{}
	for _, m := range matches {
		if m.GetString("status") != "finished" {
			continue
		}
		ko := m.GetDateTime("kickoff").Time()
		if !ko.Before(since) && ko.Before(now) {
			finalized[m.Id] = true
		}
	}
	if len(finalized) == 0 {
		return nil // nothing happened — no empty recaps
	}

	cfgID := r.defaultConfigID()
	if cfgID == "" {
		return fmt.Errorf("no default scoring config")
	}
	ncol, err := r.notificationsCol()
	if err != nil {
		return err
	}
	dateKey := now.Format("2006-01-02")
	boards := map[string][]scoring.Row{} // league id -> rows (cached this pass)

	for _, u := range recipients {
		// Only recap users who actually participate (have at least one tip).
		if n, _ := r.app.CountRecords("tips", dbx.HashExp{"user": u.Id}); n == 0 {
			continue
		}

		gained, total := r.userPoints(u.Id, cfgID, finalized)
		ranks := r.userRanks(u.Id, boards)

		data := tplData{
			Finalized:    len(finalized),
			PointsGained: gained,
			Total:        total,
			Ranks:        ranks,
			CTAText:      "See the leaderboard",
			CTAUrl:       base.url + "/leagues",
		}
		r.dispatch(ctx, res, ncol, u, "results_recap", "results_recap:"+u.Id+":"+dateKey, data)
	}
	return nil
}
