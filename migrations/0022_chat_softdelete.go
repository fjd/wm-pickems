package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Soft-delete for league chat. Deleting a message clears its live `text`
// (so members — and the realtime payload — only ever see "message deleted")
// and stashes the original in `origText`, a Hidden field excluded from API /
// realtime output for everyone but the backend, which serves it to app-admins
// through the moderation endpoint. `text` becomes optional so it can be blanked.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		col, err := app.FindCollectionByNameOrId("league_messages")
		if err != nil {
			return err
		}
		if f, ok := col.Fields.GetByName("text").(*core.TextField); ok {
			f.Required = false // allow blanking on soft-delete (create still validates)
		}
		if col.Fields.GetByName("deleted") == nil {
			col.Fields.Add(&core.BoolField{Name: "deleted"})
		}
		if col.Fields.GetByName("deletedBy") == nil {
			col.Fields.Add(&core.RelationField{Name: "deletedBy", CollectionId: users.Id, MaxSelect: 1})
		}
		if col.Fields.GetByName("deletedAt") == nil {
			col.Fields.Add(&core.DateField{Name: "deletedAt"})
		}
		if col.Fields.GetByName("origText") == nil {
			col.Fields.Add(&core.TextField{Name: "origText", Max: 2000, Hidden: true})
		}
		return app.Save(col)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("league_messages")
		if err != nil {
			return err
		}
		col.Fields.RemoveByName("deleted")
		col.Fields.RemoveByName("deletedBy")
		col.Fields.RemoveByName("deletedAt")
		col.Fields.RemoveByName("origText")
		if f, ok := col.Fields.GetByName("text").(*core.TextField); ok {
			f.Required = true
		}
		return app.Save(col)
	})
}
