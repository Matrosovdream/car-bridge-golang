// Package sendgrid implements comms.EmailSender using the SendGrid v3 API.
package sendgrid

import (
	"context"
	"fmt"
	"net/http"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/comms"
)

// Client sends transactional email through the SendGrid v3 Mail Send API.
type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

// New builds a sendgrid client from config and a shared base HTTP client.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	return &Client{base: base, cfg: cfg}
}

// compile-time check that Client satisfies the port.
var _ comms.EmailSender = (*Client)(nil)

// sendRequest is the SendGrid /v3/mail/send payload.
// See https://docs.sendgrid.com/api-reference/mail-send/mail-send
type sendRequest struct {
	Personalizations []personalization `json:"personalizations"`
	From             address           `json:"from"`
	Subject          string            `json:"subject"`
	Content          []content         `json:"content"`
}

type personalization struct {
	To []address `json:"to"`
}

type address struct {
	Email string `json:"email"`
}

type content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Send delivers a single email via SendGrid. A successful send returns 202 with
// an empty body; SendGrid auth is Bearer <API key>, which the shared client sets
// from cfg.APIKey.
func (c *Client) Send(ctx context.Context, email comms.Email) error {

	if email.From == "" {
		return fmt.Errorf("sendgrid: email.From is required")
	}
	if email.HTML == "" && email.Text == "" {
		return fmt.Errorf("sendgrid: email needs HTML or Text body")
	}

	// SendGrid requires text/plain before text/html.
	var body []content
	if email.Text != "" {
		body = append(body, content{Type: "text/plain", Value: email.Text})
	}
	if email.HTML != "" {
		body = append(body, content{Type: "text/html", Value: email.HTML})
	}

	req := sendRequest{
		Personalizations: []personalization{{To: []address{{Email: email.To}}}},
		From:             address{Email: email.From},
		Subject:          email.Subject,
		Content:          body,
	}

	return c.base.DoJSON(
		ctx,
		http.MethodPost,
		c.base.URL("/v3/mail/send"),
		req, nil, nil,
	)

}
