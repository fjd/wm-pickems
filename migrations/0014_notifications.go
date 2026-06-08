package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Notifications layer: a per-send ledger that doubles as the idempotency table
// for the notify cron, plus a per-user `notifyPrefs` JSON field on users.
//
// The ledger's unique (dedupKey, channel) index is what makes the scheduler
// safe to run every few minutes — inserting a row that already exists fails, and
// that failure IS the "already notified" signal. `providerMessageId` is stored
// so a future delivery-webhook (Mailjet Event API) can join back onto the row
// without any schema change.
const nNotifications = "notifications"

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// ---- notifications (send ledger / dedup) ----
		if _, err := app.FindCollectionByNameOrId(nNotifications); err != nil {
			n := core.NewBaseCollection(nNotifications)
			// Admin/superuser only — there's no end-user need to list the ledger.
			n.Fields.Add(&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true})
			n.Fields.Add(&core.TextField{Name: "event", Required: true, Max: 64})
			n.Fields.Add(&core.TextField{Name: "dedupKey", Required: true, Max: 200})
			n.Fields.Add(&core.SelectField{Name: "channel", Required: true, MaxSelect: 1, Values: []string{"email", "push"}})
			n.Fields.Add(&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"queued", "sent", "failed", "skipped"}})
			n.Fields.Add(&core.TextField{Name: "providerMessageId", Max: 200})
			n.Fields.Add(&core.TextField{Name: "providerStatus", Max: 64})
			n.Fields.Add(&core.TextField{Name: "error", Max: 1000})
			n.Fields.Add(&core.DateField{Name: "sentAt"})
			n.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
			n.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
			// Idempotency: one row per (dedupKey, channel).
			n.AddIndex("idx_notifications_dedup", true, "dedupKey, channel", "")
			n.AddIndex("idx_notifications_user", false, "user", "")
			if err := app.Save(n); err != nil {
				return err
			}
		}

		// ---- users.notifyPrefs (per-event channel toggles) ----
		// Absent/empty is treated as "all email events on" by the notify code,
		// so existing users need no backfill. Not a protected field, so users
		// may edit their own prefs through the public API.
		if users.Fields.GetByName("notifyPrefs") == nil {
			users.Fields.Add(&core.JSONField{Name: "notifyPrefs", MaxSize: 4000})
			if err := app.Save(users); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		if c, err := app.FindCollectionByNameOrId(nNotifications); err == nil {
			if err := app.Delete(c); err != nil {
				return err
			}
		}
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.RemoveByName("notifyPrefs")
		return app.Save(users)
	})
}
