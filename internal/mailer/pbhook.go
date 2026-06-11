package mailer

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterSystemMail routes PocketBase's own transactional emails (account
// verification, password reset, email change, OTP) through the given Sender,
// so the one configured provider manages every outgoing email and PB's SMTP
// settings don't need to be configured at all. Skipped for the smtp provider,
// which already IS PocketBase's native mail path (intercepting would recurse).
//
// brand inlines (e.g. the logo mark) are attached to any system email whose
// HTML references them via cid:<ContentID> — the branded templates set by the
// migrations rely on this, since PB templates carry HTML only.
func RegisterSystemMail(app core.App, s Sender, brand ...Inline) {
	if s.Name() == "smtp" {
		return
	}
	app.OnMailerSend().BindFunc(func(e *core.MailerEvent) error {
		// The Sender abstraction has no regular-attachment support. PB system
		// emails never carry any, but if one ever does, fall back to the
		// native mailer rather than silently dropping the file.
		if len(e.Message.Attachments) > 0 {
			return e.Next()
		}
		msg := Message{
			Subject: e.Message.Subject,
			HTML:    e.Message.HTML,
			Text:    e.Message.Text,
		}
		for _, in := range brand {
			if strings.Contains(e.Message.HTML, "cid:"+in.ContentID) {
				msg.Inlines = append(msg.Inlines, in)
			}
		}
		for name, rd := range e.Message.InlineAttachments {
			data, err := io.ReadAll(rd)
			if err != nil {
				return e.Next()
			}
			msg.Inlines = append(msg.Inlines, Inline{
				ContentID:   name,
				Filename:    name,
				ContentType: http.DetectContentType(data),
				Data:        data,
			})
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		for _, to := range e.Message.To {
			m := msg
			m.ToEmail = to.Address
			m.ToName = to.Name
			if _, err := s.Send(ctx, m); err != nil {
				log.Printf("[mailer] system mail to %s via %s failed: %v", to.Address, s.Name(), err)
				return err
			}
		}
		// Delivered via the provider — don't run PocketBase's native mailer.
		return nil
	})
}
