package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// mailjetSendURL is Mailjet's transactional Send API v3.1 endpoint. We hit it
// directly over HTTP (Basic auth = API key : secret) to avoid pulling in the
// Mailjet SDK — the payload is tiny and stable.
const mailjetSendURL = "https://api.mailjet.com/v3.1/send"

type mailjet struct {
	from   from
	apiKey string
	secret string
	http   *http.Client
}

func newMailjet(f from) *mailjet {
	return &mailjet{
		from:   f,
		apiKey: os.Getenv("MAILJET_API_KEY"),
		secret: os.Getenv("MAILJET_SECRET"),
		http:   &http.Client{Timeout: 20 * time.Second},
	}
}

func (m *mailjet) Name() string { return "mailjet" }

// mjRequest / mjMessage mirror the subset of the Send v3.1 schema we use.
type mjRequest struct {
	Messages []mjMessage `json:"Messages"`
}

type mjAddr struct {
	Email string `json:"Email"`
	Name  string `json:"Name,omitempty"`
}

type mjMessage struct {
	From     mjAddr   `json:"From"`
	To       []mjAddr `json:"To"`
	Subject  string   `json:"Subject"`
	TextPart string   `json:"TextPart,omitempty"`
	HTMLPart string   `json:"HTMLPart,omitempty"`
}

// mjResponse captures the per-message result; MessageID is what we persist for
// status tracking and a possible future webhook join.
type mjResponse struct {
	Messages []struct {
		Status string `json:"Status"`
		Errors []struct {
			ErrorMessage string `json:"ErrorMessage"`
		} `json:"Errors"`
		To []struct {
			MessageID   string `json:"MessageID"`
			MessageUUID string `json:"MessageUUID"`
		} `json:"To"`
	} `json:"Messages"`
}

func (m *mailjet) Send(ctx context.Context, msg Message) (string, error) {
	if m.from.email == "" {
		return "", fmt.Errorf("mailjet: no sender address (set MAIL_FROM or PocketBase sender)")
	}
	payload := mjRequest{Messages: []mjMessage{{
		From:     mjAddr{Email: m.from.email, Name: m.from.name},
		To:       []mjAddr{{Email: msg.ToEmail, Name: msg.ToName}},
		Subject:  msg.Subject,
		TextPart: msg.Text,
		HTMLPart: msg.HTML,
	}}}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mailjetSendURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(m.apiKey, m.secret)

	resp, err := m.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var mr mjResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		// Decode failure with a bad status is still an error worth surfacing.
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("mailjet: status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("mailjet: decode response: %w", err)
	}
	if len(mr.Messages) == 0 {
		return "", fmt.Errorf("mailjet: empty response (status %d)", resp.StatusCode)
	}
	first := mr.Messages[0]
	if first.Status != "success" {
		if len(first.Errors) > 0 {
			return "", fmt.Errorf("mailjet: %s", first.Errors[0].ErrorMessage)
		}
		return "", fmt.Errorf("mailjet: status %q (http %d)", first.Status, resp.StatusCode)
	}
	if len(first.To) > 0 {
		return first.To[0].MessageID, nil
	}
	return "", nil
}
