// Package sync keeps the matches collection up to date: a cron job pulls
// results from API-Football (one request per run), a superuser endpoint
// forces a refresh, and another superuser endpoint applies manual results
// when the provider is wrong or no API key is configured.
package sync

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/floholz/wm-pickems/internal/football"
)

// cronExpr runs the sync every 30 minutes => max 48 requests/day, comfortably
// under the API-Football free tier (100/day).
const cronExpr = "*/30 * * * *"

// nameAliases maps API-Football names that differ from the openfootball seed
// names to the seeded team name.
var nameAliases = map[string]string{
	football.NormalizeName("Korea Republic"): football.NormalizeName("South Korea"),
	football.NormalizeName("Czechia"):        football.NormalizeName("Czech Republic"),
	football.NormalizeName("USA"):            football.NormalizeName("United States"),
	football.NormalizeName("IR Iran"):        football.NormalizeName("Iran"),
}

func canonName(s string) string {
	n := football.NormalizeName(s)
	if a, ok := nameAliases[n]; ok {
		return a
	}
	return n
}

// Register wires the cron job (if an API key is set) and the HTTP endpoints.
// Called from the OnServe hook.
func Register(app core.App, se *core.ServeEvent) {
	key := os.Getenv("API_FOOTBALL_KEY")

	if key != "" {
		client := football.New(key)
		app.Cron().MustAdd("football-sync", cronExpr, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := SyncOnce(ctx, app, client); err != nil {
				log.Printf("[sync] %v", err)
			}
		})
		log.Printf("[sync] API-Football auto-sync enabled (%s)", cronExpr)
	} else {
		log.Printf("[sync] API_FOOTBALL_KEY not set: auto-sync disabled, manual override only")
	}

	// Force a sync now (superuser).
	se.Router.POST("/api/sync/refresh", func(e *core.RequestEvent) error {
		if key == "" {
			return e.JSON(400, map[string]string{"error": "API_FOOTBALL_KEY not configured"})
		}
		ctx, cancel := context.WithTimeout(e.Request.Context(), 30*time.Second)
		defer cancel()
		if err := SyncOnce(ctx, app, football.New(key)); err != nil {
			return e.JSON(500, map[string]string{"error": err.Error()})
		}
		return e.JSON(200, map[string]string{"status": "ok"})
	}).Bind(apis.RequireSuperuserAuth())

	// Manual result override (superuser). Body: ftHome,ftAway,etHome,etAway,
	// penHome,penAway (ints, et/pen optional) and status.
	se.Router.POST("/api/admin/matches/{id}/result", func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		rec, err := app.FindRecordById("matches", id)
		if err != nil {
			return e.JSON(404, map[string]string{"error": "match not found"})
		}
		var body struct {
			FTHome, FTAway   *int
			ETHome, ETAway   *int
			PenHome, PenAway *int
			Status           string
		}
		if err := e.BindBody(&body); err != nil {
			return e.JSON(400, map[string]string{"error": err.Error()})
		}
		applyResult(rec, body.Status, body.FTHome, body.FTAway, body.ETHome, body.ETAway, body.PenHome, body.PenAway)
		if err := app.Save(rec); err != nil {
			return e.JSON(500, map[string]string{"error": err.Error()})
		}
		if err := ResolveBracket(app); err != nil {
			log.Printf("[sync] resolve after manual override: %v", err)
		}
		return e.JSON(200, map[string]any{"status": "ok", "id": rec.Id})
	}).Bind(apis.RequireSuperuserAuth())
}

// SyncOnce pulls all fixtures once and updates matched records.
func SyncOnce(ctx context.Context, app core.App, client *football.Client) error {
	fixtures, err := client.Fixtures(ctx)
	if err != nil {
		return fmt.Errorf("fetch fixtures: %w", err)
	}

	matches, err := app.FindRecordsByFilter("matches", "id != ''", "kickoff", 0, 0)
	if err != nil {
		return fmt.Errorf("load matches: %w", err)
	}

	// Index our matches by the normalized team-name pair (group stage) so we
	// can line them up with provider fixtures regardless of fixture ids.
	teamName := map[string]string{} // teamId -> normalized name
	teams, _ := app.FindRecordsByFilter("teams", "id != ''", "", 0, 0)
	for _, t := range teams {
		teamName[t.Id] = canonName(t.GetString("name"))
	}

	byPair := map[string]*core.Record{}
	for _, mrec := range matches {
		h := teamName[mrec.GetString("homeTeam")]
		a := teamName[mrec.GetString("awayTeam")]
		if h != "" && a != "" {
			byPair[h+"|"+a] = mrec
		}
	}

	updated := 0
	for _, f := range fixtures {
		key := canonName(f.HomeName) + "|" + canonName(f.AwayName)
		rec, ok := byPair[key]
		if !ok {
			// KO matches resolve via ResolveBracket; unmatched group names
			// usually mean an alias is missing — logged, not fatal.
			continue
		}
		status := "scheduled"
		switch {
		case f.Finished():
			status = "finished"
		case f.Live():
			status = "live"
		}
		applyResult(rec, status, f.FTHome, f.FTAway, f.ETHome, f.ETAway, f.PenHome, f.PenAway)
		if app.Save(rec) == nil {
			updated++
		}
	}

	if err := ResolveBracket(app); err != nil {
		log.Printf("[sync] resolve bracket: %v", err)
	}
	log.Printf("[sync] fixtures=%d updated=%d", len(fixtures), updated)
	return nil
}

func ip(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// applyResult writes scores/status onto a match record and, for knockout
// matches, derives the advancer (ET > penalties > regulation).
func applyResult(rec *core.Record, status string, ftH, ftA, etH, etA, penH, penA *int) {
	if status != "" {
		rec.Set("status", status)
	}
	if ftH != nil {
		rec.Set("ftHome", *ftH)
	}
	if ftA != nil {
		rec.Set("ftAway", *ftA)
	}
	rec.Set("etHome", ip(etH))
	rec.Set("etAway", ip(etA))
	rec.Set("penHome", ip(penH))
	rec.Set("penAway", ip(penA))

	finished := rec.GetString("status") == "finished"
	if finished {
		rec.Set("finalizedAt", time.Now().UTC())
	}

	if rec.GetString("stage") == "group" || !finished {
		return
	}
	// Knockout advancer resolution.
	home := rec.GetString("homeTeam")
	away := rec.GetString("awayTeam")
	switch {
	case penH != nil && penA != nil && *penH != *penA:
		if *penH > *penA {
			rec.Set("penWinner", home)
			rec.Set("advancer", home)
		} else {
			rec.Set("penWinner", away)
			rec.Set("advancer", away)
		}
	case etH != nil && etA != nil && *etH != *etA:
		if *etH > *etA {
			rec.Set("advancer", home)
		} else {
			rec.Set("advancer", away)
		}
	case ftH != nil && ftA != nil && *ftH != *ftA:
		if *ftH > *ftA {
			rec.Set("advancer", home)
		} else {
			rec.Set("advancer", away)
		}
	}
}
