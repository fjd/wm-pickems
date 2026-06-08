package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Rename the already-seeded team display name "Turkey" -> "Türkiye". Fresh
// installs get the override straight from the seeder (internal/seed displayNames);
// this fixes databases seeded before that change. Matching is unaffected: the
// stored match extIds (slug of the openfootball name) and the API-Football name
// alias both stay keyed off "Turkey", so only the user-facing label changes.
//
// On a fresh DB this runs before the seeder populates teams and simply no-ops.
func init() {
	m.Register(func(app core.App) error {
		rec, err := app.FindFirstRecordByFilter("teams",
			"name = {:n}", map[string]any{"n": "Turkey"})
		if err != nil {
			return nil // not seeded yet (or already renamed) — nothing to do
		}
		rec.Set("name", "Türkiye")
		return app.Save(rec)
	}, func(app core.App) error {
		rec, err := app.FindFirstRecordByFilter("teams",
			"name = {:n}", map[string]any{"n": "Türkiye"})
		if err != nil {
			return nil
		}
		rec.Set("name", "Turkey")
		return app.Save(rec)
	})
}
