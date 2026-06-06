package notify

import (
	"testing"
	"time"
)

func TestInLeadWindow(t *testing.T) {
	now := time.Date(2026, 6, 11, 6, 0, 0, 0, time.UTC)
	lead := 12 * time.Hour
	tests := []struct {
		name  string
		start time.Time
		want  bool
	}{
		{"in the past", now.Add(-time.Hour), false},
		{"now exactly (not future)", now, false},
		{"just inside future", now.Add(time.Minute), true},
		{"mid-window", now.Add(6 * time.Hour), true},
		{"at the lead edge", now.Add(12 * time.Hour), true},
		{"just past the lead edge", now.Add(12*time.Hour + time.Minute), false},
		{"far future", now.Add(48 * time.Hour), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := inLeadWindow(now, tc.start, lead); got != tc.want {
				t.Fatalf("inLeadWindow(%v) = %v, want %v", tc.start.Sub(now), got, tc.want)
			}
		})
	}
}

func TestHumanizeDur(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Minute, "less than an hour"},
		{59 * time.Minute, "less than an hour"},
		{time.Hour, "1 hours"},
		{90 * time.Minute, "2 hours"}, // rounds to nearest hour
		{11*time.Hour + 40*time.Minute, "12 hours"},
		{47 * time.Hour, "47 hours"},
		{48 * time.Hour, "2 days"},
		{60 * time.Hour, "3 days"}, // 2.5d rounds up
	}
	for _, tc := range tests {
		if got := humanizeDur(tc.d); got != tc.want {
			t.Errorf("humanizeDur(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestPrefEnabledFromRaw(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		event string
		want  bool
	}{
		{"empty prefs default on", "", "tips_reminder", true},
		{"invalid json default on", "{not json", "tips_reminder", true},
		{"event absent default on", `{"stage_starting":{"email":true}}`, "tips_reminder", true},
		{"email key absent default on", `{"tips_reminder":{}}`, "tips_reminder", true},
		{"explicitly on", `{"tips_reminder":{"email":true}}`, "tips_reminder", true},
		{"explicitly off", `{"tips_reminder":{"email":false}}`, "tips_reminder", false},
		{"off for a different event only", `{"results_recap":{"email":false}}`, "tips_reminder", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := prefEnabledFromRaw(tc.raw, tc.event); got != tc.want {
				t.Fatalf("prefEnabledFromRaw(%q,%q) = %v, want %v", tc.raw, tc.event, got, tc.want)
			}
		})
	}
}

func intp(v int) *int { return &v }

func TestApplyConfigDefaults(t *testing.T) {
	tests := []struct {
		name        string
		in          storedConfig
		wantLead    int
		wantRecapHr int
	}{
		{"all unset -> defaults", storedConfig{}, defaultLeadHours, defaultRecapHourUTC},
		{"lead set", storedConfig{LeadHours: intp(6)}, 6, defaultRecapHourUTC},
		{"lead zero ignored", storedConfig{LeadHours: intp(0)}, defaultLeadHours, defaultRecapHourUTC},
		{"lead negative ignored", storedConfig{LeadHours: intp(-3)}, defaultLeadHours, defaultRecapHourUTC},
		{"recap midnight honored", storedConfig{RecapHourUTC: intp(0)}, defaultLeadHours, 0},
		{"recap set", storedConfig{RecapHourUTC: intp(20)}, defaultLeadHours, 20},
		{"recap out of range ignored", storedConfig{RecapHourUTC: intp(25)}, defaultLeadHours, defaultRecapHourUTC},
		{"both set", storedConfig{LeadHours: intp(24), RecapHourUTC: intp(9)}, 24, 9},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := applyConfigDefaults(tc.in)
			if got.LeadHours != tc.wantLead || got.RecapHourUTC != tc.wantRecapHr {
				t.Fatalf("applyConfigDefaults(%+v) = %+v, want lead=%d recap=%d",
					tc.in, got, tc.wantLead, tc.wantRecapHr)
			}
		})
	}
}

func TestRenderAllTemplates(t *testing.T) {
	// Each event template must parse and execute against tplData without error,
	// and produce a non-empty subject + bodies.
	events := []string{"stage_starting", "forecast_reminder", "tips_reminder", "results_recap"}
	data := tplData{
		AppName:      "WM Pickems",
		SettingsUrl:  "https://example.test/settings",
		CTAText:      "Go",
		CTAUrl:       "https://example.test/tips",
		StageName:    "Round of 32",
		StartsIn:     "12 hours",
		WhenText:     "Sat, Jun 28 · 18:00 UTC",
		Count:        2,
		Matches:      []matchLine{{Home: "Brazil", Away: "Spain", WhenText: "soon"}},
		Finalized:    3,
		PointsGained: 7,
		Total:        42,
		Ranks:        []rankLine{{League: "Friends", Rank: 1, Of: 8}},
	}
	for _, ev := range events {
		t.Run(ev, func(t *testing.T) {
			subject, html, text, err := render(ev, data)
			if err != nil {
				t.Fatalf("render(%s): %v", ev, err)
			}
			if subject == "" || html == "" || text == "" {
				t.Fatalf("render(%s): empty output (subj=%q htmlLen=%d textLen=%d)",
					ev, subject, len(html), len(text))
			}
		})
	}
}
