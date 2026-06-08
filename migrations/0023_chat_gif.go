package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Chat GIFs (v3): a message can be text OR a GIF (a hosted URL from the KLIPY
// proxy — we don't store the file). `gif` holds the live URL; on soft-delete it's
// cleared into the hidden `origGif` (mirrors text/origText) so members and the
// realtime payload only see "message deleted" while admins keep the original.
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("league_messages")
		if err != nil {
			return err
		}
		if col.Fields.GetByName("gif") == nil {
			col.Fields.Add(&core.TextField{Name: "gif", Max: 1000})
		}
		if col.Fields.GetByName("origGif") == nil {
			col.Fields.Add(&core.TextField{Name: "origGif", Max: 1000, Hidden: true})
		}
		return app.Save(col)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("league_messages")
		if err != nil {
			return err
		}
		col.Fields.RemoveByName("gif")
		col.Fields.RemoveByName("origGif")
		return app.Save(col)
	})
}
