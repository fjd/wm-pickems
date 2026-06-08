// Package mailer is a provider-agnostic email transport. The notify layer talks
// to the Sender interface only; the concrete provider (Mailjet, generic SMTP, or
// a dev log sink) is chosen at boot from the environment, so swapping providers
// never touches calling code.
package mailer

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// Inline is an image embedded in the message and referenced from the HTML via
// cid:<ContentID> — so it renders without a remote fetch.
type Inline struct {
	ContentID   string
	Filename    string
	ContentType string
	Data        []byte
}

// Message is one email to one recipient.
type Message struct {
	ToEmail string
	ToName  string
	Subject string
	HTML    string
	Text    string
	Inlines []Inline
}

// Sender delivers a single message and returns the provider's message id (empty
// if the provider doesn't expose one, e.g. plain SMTP). A non-nil error means
// the provider rejected or failed to accept the message.
type Sender interface {
	Name() string
	Send(ctx context.Context, m Message) (messageID string, err error)
}

// from holds the resolved sender identity.
type from struct {
	email string
	name  string
}

// resolveFrom prefers MAIL_FROM / MAIL_FROM_NAME, falling back to PocketBase's
// configured sender (Settings → Meta), so a single source of truth still works.
func resolveFrom(app core.App) from {
	f := from{
		email: os.Getenv("MAIL_FROM"),
		name:  os.Getenv("MAIL_FROM_NAME"),
	}
	meta := app.Settings().Meta
	if f.email == "" {
		f.email = meta.SenderAddress
	}
	if f.name == "" {
		f.name = meta.SenderName
	}
	return f
}

// Pick returns the active Sender, selected by MAIL_PROVIDER
// (mailjet|smtp|log), else auto-detected: Mailjet when its keys are present,
// generic SMTP when PocketBase SMTP is enabled, otherwise the log sink. It
// always returns a usable Sender (never nil) and logs the choice.
func Pick(app core.App) Sender {
	f := resolveFrom(app)
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("MAIL_PROVIDER")))

	mailjetReady := os.Getenv("MAILJET_API_KEY") != "" && os.Getenv("MAILJET_SECRET") != ""
	smtpReady := app.Settings().SMTP.Enabled

	switch mode {
	case "mailjet":
		s := newMailjet(f)
		log.Printf("[mailer] using mailjet (forced)")
		return s
	case "smtp":
		log.Printf("[mailer] using smtp (forced)")
		return newSMTP(app, f)
	case "log":
		log.Printf("[mailer] using log sink (forced) — emails are not delivered")
		return newLog(f)
	}

	switch {
	case mailjetReady:
		log.Printf("[mailer] using mailjet")
		return newMailjet(f)
	case smtpReady:
		log.Printf("[mailer] using smtp (PocketBase settings)")
		return newSMTP(app, f)
	default:
		log.Printf("[mailer] no provider configured — using log sink (emails are not delivered)")
		return newLog(f)
	}
}
