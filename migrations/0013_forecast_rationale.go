package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Add an optional `rationale` JSON blob to forecasts — the reasoning behind the
// picks, keyed to mirror the existing fields: {"groups":{letter:text},
// "bracket":{bracketKey:text}}. Bot players (Claude with BOT_RATIONALE on) fill
// it; human forecasts leave it empty/null. Like the rest of the forecast it's
// hidden until the forecast locks, then revealed — for displaying "why" later.
func init() {
	m.Register(func(app core.App) error {
		fc, err := app.FindCollectionByNameOrId("forecasts")
		if err != nil {
			return err
		}
		if fc.Fields.GetByName("rationale") == nil {
			fc.Fields.Add(&core.JSONField{Name: "rationale", MaxSize: 60000})
		}
		return app.Save(fc)
	}, func(app core.App) error {
		fc, err := app.FindCollectionByNameOrId("forecasts")
		if err != nil {
			return err
		}
		fc.Fields.RemoveByName("rationale")
		return app.Save(fc)
	})
}
