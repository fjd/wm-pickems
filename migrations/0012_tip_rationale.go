package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Add an optional free-text `rationale` to tips — the reasoning behind a pick.
// Bot players (e.g. the Claude bot with BOT_RATIONALE on) fill it; human tips
// leave it empty. It rides on the tip record, so it inherits tip visibility
// (hidden until kickoff, revealed after) — meant for displaying "why" later.
func init() {
	m.Register(func(app core.App) error {
		tips, err := app.FindCollectionByNameOrId("tips")
		if err != nil {
			return err
		}
		if tips.Fields.GetByName("rationale") == nil {
			tips.Fields.Add(&core.TextField{Name: "rationale", Max: 2000})
		}
		return app.Save(tips)
	}, func(app core.App) error {
		tips, err := app.FindCollectionByNameOrId("tips")
		if err != nil {
			return err
		}
		tips.Fields.RemoveByName("rationale")
		return app.Save(tips)
	})
}
