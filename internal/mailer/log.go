package mailer

import (
	"context"
	"fmt"
	"log"
)

// logSender is the no-op fallback used in development or when no real provider
// is configured. It logs a one-line summary (not the full body) and returns a
// synthetic message id so the rest of the pipeline behaves as if a send
// succeeded — handy for exercising the notify flow without delivering mail.
type logSender struct {
	from from
	seq  int
}

func newLog(f from) *logSender { return &logSender{from: f} }

func (l *logSender) Name() string { return "log" }

func (l *logSender) Send(_ context.Context, m Message) (string, error) {
	l.seq++
	log.Printf("[mailer:log] to=%q subject=%q (not actually sent)", m.ToEmail, m.Subject)
	return fmt.Sprintf("log-%d", l.seq), nil
}
