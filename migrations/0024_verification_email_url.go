package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Point the verification email at the SPA route so users land on a branded
// /confirm-verification/<token> page rather than PocketBase's admin UI at
// /_/#/auth/confirm-verification/<token>. Same pattern as the password-reset
// rewrite in 0008 — the token + behavior is identical, only the link differs.
func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		before := users.VerificationTemplate.Body
		after := strings.ReplaceAll(before,
			"/_/#/auth/confirm-verification/",
			"/confirm-verification/",
		)
		if before == after {
			return nil
		}
		users.VerificationTemplate.Body = after
		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		before := users.VerificationTemplate.Body
		after := strings.ReplaceAll(before,
			"/confirm-verification/",
			"/_/#/auth/confirm-verification/",
		)
		if before == after {
			return nil
		}
		users.VerificationTemplate.Body = after
		return app.Save(users)
	})
}
