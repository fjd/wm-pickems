package notify

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

//go:embed templates/*.html templates/*.txt
var tmplFS embed.FS

// matchLine is one fixture row in the tips-reminder digest.
type matchLine struct {
	Home     string
	Away     string
	WhenText string
}

// rankLine is one league standing in the results recap.
type rankLine struct {
	League string
	Rank   int
	Of     int
}

// tplData is the union of fields every template might reference. Unused fields
// for a given event are simply left zero. Common fields (AppName, SettingsUrl,
// CTA*) are filled by the dispatcher; the rest by each detector.
type tplData struct {
	AppName     string
	SettingsUrl string
	CTAText     string
	CTAUrl      string

	StageName string
	StartsIn  string
	WhenText  string

	Count   int
	Matches []matchLine

	Finalized    int
	PointsGained int
	Total        int
	Ranks        []rankLine
}

// render produces the subject, HTML body, and text body for an event by
// combining the shared layout with the event's own partials.
func render(event string, data tplData) (subject, html, text string, err error) {
	ht, err := htmltemplate.New("").ParseFS(tmplFS,
		"templates/layout.html", "templates/"+event+".html")
	if err != nil {
		return "", "", "", fmt.Errorf("parse html %s: %w", event, err)
	}
	var sb, hb bytes.Buffer
	if err := ht.ExecuteTemplate(&sb, "subject", data); err != nil {
		return "", "", "", fmt.Errorf("subject %s: %w", event, err)
	}
	if err := ht.ExecuteTemplate(&hb, "layout", data); err != nil {
		return "", "", "", fmt.Errorf("html %s: %w", event, err)
	}

	tt, err := texttemplate.New(event+".txt").ParseFS(tmplFS, "templates/"+event+".txt")
	if err != nil {
		return "", "", "", fmt.Errorf("parse text %s: %w", event, err)
	}
	var tb bytes.Buffer
	if err := tt.Execute(&tb, data); err != nil {
		return "", "", "", fmt.Errorf("text %s: %w", event, err)
	}

	return sb.String(), hb.String(), tb.String(), nil
}
