package notify

import "github.com/pocketbase/pocketbase/core"

// config holds the resolved (defaults-applied) notify settings.
type config struct {
	LeadHours    int // reminder lead time before a deadline
	RecapHourUTC int // hour (UTC) the daily results recap fires
}

// storedConfig is the on-disk shape (app_meta "notify_config"). Pointers let us
// tell "unset" apart from a legitimate zero (e.g. recapHourUTC = 0 = midnight).
type storedConfig struct {
	LeadHours    *int `json:"leadHours"`
	RecapHourUTC *int `json:"recapHourUTC"`
}

const (
	defaultLeadHours    = 12
	defaultRecapHourUTC = 8
	metaKey             = "notify_config"
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

// applyConfigDefaults fills unset/out-of-range fields with the defaults. Pure,
// so it's unit-testable without an app.
func applyConfigDefaults(s storedConfig) config {
	c := config{LeadHours: defaultLeadHours, RecapHourUTC: defaultRecapHourUTC}
	if s.LeadHours != nil && *s.LeadHours > 0 {
		c.LeadHours = *s.LeadHours
	}
	if s.RecapHourUTC != nil && *s.RecapHourUTC >= 0 && *s.RecapHourUTC <= 23 {
		c.RecapHourUTC = *s.RecapHourUTC
	}
	return c
}
