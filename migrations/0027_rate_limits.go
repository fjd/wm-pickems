package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Enable PocketBase's built-in per-IP rate limiter (it ships disabled) using
// the stock rules already present in settings (*:auth 2req/3s, *:create
// 20req/5s, /api/ 300req/10s, …) — brute-force protection for auth endpoints
// and a backstop for everything else.
//
// The app runs behind a TLS-terminating reverse proxy (see DEPLOY.md), so the
// real client IP arrives in X-Forwarded-For; it must be trusted or every
// request appears to come from the proxy's IP and all users share one rate
// bucket. UseLeftmostIP stays false: behind a single trusted proxy the
// rightmost entry is the proxy-observed client and can't be spoofed by a
// client-supplied header. If the app is ever exposed without a proxy, clear
// the trusted headers in Dashboard → Settings → Application.
func init() {
	m.Register(func(app core.App) error {
		s := app.Settings()
		s.RateLimits.Enabled = true
		s.TrustedProxy.Headers = []string{"X-Forwarded-For"}
		s.TrustedProxy.UseLeftmostIP = false
		return app.Save(s)
	}, func(app core.App) error {
		s := app.Settings()
		s.RateLimits.Enabled = false
		s.TrustedProxy.Headers = nil
		return app.Save(s)
	})
}
