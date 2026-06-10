package mailer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"time"
)

// resendSendURL is Resend's transactional send endpoint. Like Mailjet, we hit
// it directly over HTTP (Bearer auth = API key) to avoid pulling in the SDK —
// the payload is tiny and stable.
const resendSendURL = "https://api.resend.com/emails"

type resend struct {
	from   from
	apiKey string
	http   *http.Client
}

func newResend(f from) *resend {
	return &resend{
		from:   f,
		apiKey: os.Getenv("RESEND_API_KEY"),
		http:   &http.Client{Timeout: 20 * time.Second},
	}
}

func (r *resend) Name() string { return "resend" }

// rsRequest mirrors the subset of the POST /emails schema we use. Inline
// images go as attachments with a ContentID and are referenced from the HTML
// via cid:<id> — same model as Mailjet's InlinedAttachments.
type rsRequest struct {
	From        string         `json:"from"`
	To          []string       `json:"to"`
	Subject     string         `json:"subject"`
	HTML        string         `json:"html,omitempty"`
	Text        string         `json:"text,omitempty"`
	Attachments []rsAttachment `json:"attachments,omitempty"`
}

type rsAttachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"` // base64
	ContentType string `json:"content_type,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
}

func (r *resend) Send(ctx context.Context, msg Message) (string, error) {
	if r.from.email == "" {
		return "", fmt.Errorf("resend: no sender address (set MAIL_FROM or PocketBase sender)")
	}
	payload := rsRequest{
		// mail.Address takes care of quoting/encoding the display name in the
		// "Name <addr>" friendly format Resend expects.
		From:    (&mail.Address{Name: r.from.name, Address: r.from.email}).String(),
		To:      []string{(&mail.Address{Name: msg.ToName, Address: msg.ToEmail}).String()},
		Subject: msg.Subject,
		HTML:    msg.HTML,
		Text:    msg.Text,
	}
	for _, in := range msg.Inlines {
		payload.Attachments = append(payload.Attachments, rsAttachment{
			Filename:    in.Filename,
			Content:     base64.StdEncoding.EncodeToString(in.Data),
			ContentType: in.ContentType,
			ContentID:   in.ContentID,
		})
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendSendURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Errors come back as {"statusCode":..., "name":"...", "message":"..."}.
		var rerr struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}
		if json.NewDecoder(resp.Body).Decode(&rerr) == nil && rerr.Message != "" {
			return "", fmt.Errorf("resend: %s (http %d)", rerr.Message, resp.StatusCode)
		}
		return "", fmt.Errorf("resend: status %d", resp.StatusCode)
	}

	var rr struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return "", fmt.Errorf("resend: decode response: %w", err)
	}
	return rr.ID, nil
}
