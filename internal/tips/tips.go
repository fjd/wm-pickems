// Package tips enforces the per-match prediction rules server-side:
//   - a Tip can only be created/edited while now < match.kickoff (lock)
//   - knockout Tips are only allowed once both teams are resolved
//   - the knockout advancer is derived from the phased prediction
//   - other players' Tips are visible only AFTER kickoff and only to people
//     who share a League (the /api/tips/others/{matchId} endpoint)
package tips

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/floholz/wm-pickems/internal/clock"
)

func matchKickoff(m *core.Record) time.Time {
	return m.GetDateTime("kickoff").Time()
}

func locked(app core.App, m *core.Record) bool {
	return !clock.Now(app).Before(matchKickoff(m))
}

// bypass lets the dev bot generator insert tips for every match regardless
// of lock / knockout-resolution. Never set in production (dev-only path).
var bypass atomic.Bool

// SetBypass toggles the dev-only validation bypass.
func SetBypass(b bool) { bypass.Store(b) }

// validateAndDerive applies lock + validation and sets the derived advancer.
func validateAndDerive(app core.App, tip *core.Record) error {
	if bypass.Load() {
		return nil
	}
	match, err := app.FindRecordById("matches", tip.GetString("match"))
	if err != nil {
		return apis.NewBadRequestError("unknown match", nil)
	}
	if locked(app, match) {
		return apis.NewBadRequestError("this match is locked (kickoff passed)", nil)
	}

	ftH := tip.GetInt("ftHome")
	ftA := tip.GetInt("ftAway")
	if tip.Get("ftHome") == nil || tip.Get("ftAway") == nil {
		return apis.NewBadRequestError("full-time score is required", nil)
	}
	if ftH < 0 || ftA < 0 || ftH > 99 || ftA > 99 {
		return apis.NewBadRequestError("scores out of range", nil)
	}

	if match.GetString("stage") == "group" {
		tip.Set("etHome", 0)
		tip.Set("etAway", 0)
		tip.Set("penWinner", "")
		tip.Set("advancer", "")
		return nil
	}

	// Knockout.
	home := match.GetString("homeTeam")
	away := match.GetString("awayTeam")
	if home == "" || away == "" {
		return apis.NewBadRequestError("this matchup is not set yet", nil)
	}

	if ftH != ftA {
		if ftH > ftA {
			tip.Set("advancer", home)
		} else {
			tip.Set("advancer", away)
		}
		tip.Set("etHome", 0)
		tip.Set("etAway", 0)
		tip.Set("penWinner", "")
		return nil
	}

	// Drawn after 90' -> extra time required (cumulative >= FT).
	etH := tip.GetInt("etHome")
	etA := tip.GetInt("etAway")
	if tip.Get("etHome") == nil || tip.Get("etAway") == nil {
		return apis.NewBadRequestError("predict the score after extra time", nil)
	}
	if etH < ftH || etA < ftA {
		return apis.NewBadRequestError("extra-time score must include the 90' goals", nil)
	}
	if etH != etA {
		if etH > etA {
			tip.Set("advancer", home)
		} else {
			tip.Set("advancer", away)
		}
		tip.Set("penWinner", "")
		return nil
	}

	// Still level -> penalty winner required.
	pw := tip.GetString("penWinner")
	if pw != home && pw != away {
		return apis.NewBadRequestError("pick who wins the penalty shootout", nil)
	}
	tip.Set("advancer", pw)
	return nil
}

// Register wires the Tip validation hooks and the friends-tips endpoint.
func Register(app core.App, se *core.ServeEvent) {
	app.OnRecordCreate("tips").BindFunc(func(e *core.RecordEvent) error {
		if err := validateAndDerive(e.App, e.Record); err != nil {
			return err
		}
		return e.Next()
	})
	app.OnRecordUpdate("tips").BindFunc(func(e *core.RecordEvent) error {
		if err := validateAndDerive(e.App, e.Record); err != nil {
			return err
		}
		return e.Next()
	})
	app.OnRecordDelete("tips").BindFunc(func(e *core.RecordEvent) error {
		if m, err := e.App.FindRecordById("matches", e.Record.GetString("match")); err == nil && locked(e.App, m) {
			return apis.NewBadRequestError("this match is locked", nil)
		}
		return e.Next()
	})

	// GET /api/tips/others/{matchId} — other members' Tips, but only after
	// kickoff and only for users who share at least one League with you.
	se.Router.GET("/api/tips/others/{matchId}", func(e *core.RequestEvent) error {
		matchID := e.Request.PathValue("matchId")
		match, err := app.FindRecordById("matches", matchID)
		if err != nil {
			return apis.NewNotFoundError("match not found", nil)
		}
		if !locked(app, match) {
			// Not started: never reveal others' picks.
			return e.JSON(http.StatusOK, map[string]any{"locked": false, "tips": []any{}})
		}

		coMembers, err := sharedLeagueUserIDs(app, e.Auth.Id)
		if err != nil {
			return err
		}
		allTips, err := app.FindRecordsByFilter("tips",
			"match = {:m}", "", 0, 0, map[string]any{"m": matchID})
		if err != nil {
			return err
		}
		rows := make([]map[string]any, 0)
		for _, t := range allTips {
			uid := t.GetString("user")
			if uid == e.Auth.Id || !coMembers[uid] {
				continue
			}
			u, err := app.FindRecordById("users", uid)
			if err != nil {
				continue
			}
			rows = append(rows, map[string]any{
				"userId":    uid,
				"name":      u.GetString("name"),
				"ftHome":    t.GetInt("ftHome"),
				"ftAway":    t.GetInt("ftAway"),
				"etHome":    t.GetInt("etHome"),
				"etAway":    t.GetInt("etAway"),
				"penWinner": t.GetString("penWinner"),
				"advancer":  t.GetString("advancer"),
			})
		}
		return e.JSON(http.StatusOK, map[string]any{"locked": true, "tips": rows})
	}).Bind(apis.RequireAuth())
}

// sharedLeagueUserIDs returns the set of user ids that share at least one
// League with the given user.
func sharedLeagueUserIDs(app core.App, userID string) (map[string]bool, error) {
	mine, err := app.FindRecordsByFilter("league_members",
		"user = {:u}", "", 0, 0, map[string]any{"u": userID})
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, lm := range mine {
		peers, err := app.FindRecordsByFilter("league_members",
			"league = {:l}", "", 0, 0, map[string]any{"l": lm.GetString("league")})
		if err != nil {
			return nil, err
		}
		for _, p := range peers {
			out[p.GetString("user")] = true
		}
	}
	return out, nil
}
