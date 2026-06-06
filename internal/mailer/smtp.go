package mailer

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/pocketbase/pocketbase/core"
	pbmailer "github.com/pocketbase/pocketbase/tools/mailer"
)

// smtpSender wraps PocketBase's built-in mail client (the same transport that
// sends password-reset emails). It honours whatever SMTP settings are configured
// in the PocketBase dashboard, so it also covers "Mailjet via SMTP relay" for
// anyone who prefers that over the REST API. Plain SMTP exposes no provider
// message id, so Send returns "".
type smtpSender struct {
	app  core.App
	from from
}

func newSMTP(app core.App, f from) *smtpSender { return &smtpSender{app: app, from: f} }

func (s *smtpSender) Name() string { return "smtp" }

func (s *smtpSender) Send(ctx context.Context, m Message) (string, error) {
	if s.from.email == "" {
		return "", fmt.Errorf("smtp: no sender address (set MAIL_FROM or PocketBase sender)")
	}
	msg := &pbmailer.Message{
		From:    mail.Address{Address: s.from.email, Name: s.from.name},
		To:      []mail.Address{{Address: m.ToEmail, Name: m.ToName}},
		Subject: m.Subject,
		HTML:    m.HTML,
		Text:    m.Text,
	}
	if err := s.app.NewMailClient().Send(msg); err != nil {
		return "", err
	}
	return "", nil
}
