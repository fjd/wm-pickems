package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Brand the confirm-email-change email (same shell as 0025) and point it at
// the SPA's /confirm-email-change/<token> page instead of the PocketBase
// admin UI. Sent to the NEW address; confirming there requires the account
// password and marks the address verified.
var sysEmailChangeBody = sysEmail(`<div style="font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#c8fb50;">Account</div>
<h1 style="margin:7px 0 14px;font-size:27px;line-height:1.12;font-weight:800;color:#f3f5fb;">Confirm your new email</h1>
<p style="margin:0;font-size:15px;line-height:1.6;color:#aeb6d0;">
Click the button below to make this address the new email for your {APP_NAME} account. You'll be asked for your account password to confirm.
</p>`,
	"Confirm new email", "{APP_URL}/confirm-email-change/{TOKEN}")

func init() {
	m.Register(func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.ConfirmEmailChangeTemplate.Body = sysEmailChangeBody
		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		// Restore the PocketBase stock body (admin-UI link — no earlier
		// migration ever touched this template).
		users.ConfirmEmailChangeTemplate.Body = `<p>Hello,</p>
<p>Click on the button below to confirm your new email address.</p>
<p>
  <a class="btn" href="{APP_URL}/_/#/auth/confirm-email-change/{TOKEN}" target="_blank" rel="noopener">Confirm new email</a>
</p>
<p><i>If you didn't ask to change your email address, you can ignore this email.</i></p>
<p>
  Thanks,<br/>
  {APP_NAME} team
</p>`
		return app.Save(users)
	})
}
