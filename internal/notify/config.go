package notify

import (
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// config holds the resolved (defaults-applied) notify settings.
type config struct {
	LeadHours        int // reminder lead time before a deadline
	RecapHourUTC     int // hour (UTC) the daily results recap fires
	CountdownHourUTC int // hour (UTC) the daily pre-tournament countdown fires
	// Allowlist gates delivery to specific email addresses (lowercased) for a
	// gradual rollout. Empty = send to everyone.
	Allowlist []string
}

// storedConfig is the on-disk shape (app_meta "notify_config"). Pointers let us
// tell "unset" apart from a legitimate zero (e.g. recapHourUTC = 0 = midnight).
type storedConfig struct {
	LeadHours        *int     `json:"leadHours"`
	RecapHourUTC     *int     `json:"recapHourUTC"`
	CountdownHourUTC *int     `json:"countdownHourUTC"`
	Allowlist        []string `json:"allowlist"`
}

const (
	defaultLeadHours        = 12
	defaultRecapHourUTC     = 8
	defaultCountdownHourUTC = 9
	metaKey                 = "notify_config"
)

// readConfig loads the notify config from app_meta, applying defaults for any
// unset field. Settings are runtime-tunable from the PocketBase dashboard
// without a redeploy.
func readConfig(app core.App) config {
	rec, err := app.FindFirstRecordByFilter("app_meta",
		"key = {:k}", map[string]any{"k": metaKey})
	if err != nil {
		return applyConfigDefaults(storedConfig{})
	}
	var stored storedConfig
	if err := rec.UnmarshalJSONField("value", &stored); err != nil {
		return applyConfigDefaults(storedConfig{})
	}
	return applyConfigDefaults(stored)
}

// applyConfigDefaults fills unset/out-of-range fields with the defaults and
// resolves the allowlist (app_meta value, else the NOTIFY_ALLOWLIST env seed).
// Pure except for the env read, so the numeric defaults stay unit-testable.
func applyConfigDefaults(s storedConfig) config {
	c := config{
		LeadHours:        defaultLeadHours,
		RecapHourUTC:     defaultRecapHourUTC,
		CountdownHourUTC: defaultCountdownHourUTC,
	}
	if s.LeadHours != nil && *s.LeadHours > 0 {
		c.LeadHours = *s.LeadHours
	}
	if s.RecapHourUTC != nil && *s.RecapHourUTC >= 0 && *s.RecapHourUTC <= 23 {
		c.RecapHourUTC = *s.RecapHourUTC
	}
	if s.CountdownHourUTC != nil && *s.CountdownHourUTC >= 0 && *s.CountdownHourUTC <= 23 {
		c.CountdownHourUTC = *s.CountdownHourUTC
	}
	c.Allowlist = normalizeEmails(s.Allowlist)
	if len(c.Allowlist) == 0 {
		c.Allowlist = normalizeEmails(strings.Split(os.Getenv("NOTIFY_ALLOWLIST"), ","))
	}
	return c
}

// normalizeEmails lowercases, trims, and drops empties.
func normalizeEmails(in []string) []string {
	out := make([]string, 0, len(in))
	for _, e := range in {
		if e = strings.ToLower(strings.TrimSpace(e)); e != "" {
			out = append(out, e)
		}
	}
	return out
}
