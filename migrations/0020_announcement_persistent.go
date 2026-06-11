package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Add a `persistent` flag to announcements. A persistent announcement can't be
// dismissed by users — only collapsed to a slim ribbon (and re-expanded). It
// stays until an admin deactivates it. Absent/false keeps the default
// dismissible behaviour, so existing announcements need no backfill.
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("announcements")
		if err != nil {
			return err
		}
		if col.Fields.GetByName("persistent") == nil {
			col.Fields.Add(&core.BoolField{Name: "persistent"})
			return app.Save(col)
		}
		return nil
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("announcements")
		if err != nil {
			return err
		}
		col.Fields.RemoveByName("persistent")
		return app.Save(col)
	})
}
