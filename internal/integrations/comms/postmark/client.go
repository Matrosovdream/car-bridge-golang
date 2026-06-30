// Package postmark implements comms.EmailSender using the Postmark API.
package postmark

import (
	"context"
	"fmt"
	"net/http"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/comms"
)

// authHeader is Postmark's server-token header. Postmark authenticates the
// transactional /email endpoint with this rather than a Bearer token.
const authHeader = "X-Postmark-Server-Token"

// Client sends transactional email through the Postmark API.
type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

// New builds a postmark client from config and a shared base HTTP client.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	return &Client{base: base, cfg: cfg}
}

// compile-time check that Client satisfies the port.
var _ comms.EmailSender = (*Client)(nil)

// sendRequest is the Postmark /email payload.
// See https://postmarkapp.com/developer/api/email-api#send-a-single-email
type sendRequest struct {
	From          string `json:"From"`
	To            string `json:"To"`
	Subject       string `json:"Subject"`
	HtmlBody      string `json:"HtmlBody,omitempty"`
	TextBody      string `json:"TextBody,omitempty"`
	MessageStream string `json:"MessageStream,omitempty"`
}

// sendResponse is the Postmark /email reply. ErrorCode 0 means accepted; any
// other value is a delivery rejection even when the HTTP status is 2xx.
type sendResponse struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
	MessageID string `json:"MessageID"`
}

// Send delivers a single email via Postmark's transactional stream.
func (c *Client) Send(ctx context.Context, email comms.Email) error {

	if email.From == "" {
		return fmt.Errorf("postmark: email.From is required")
	}
	if email.HTML == "" && email.Text == "" {
		return fmt.Errorf("postmark: email needs HTML or Text body")
	}

	req := sendRequest{
		From:          email.From,
		To:            email.To,
		Subject:       email.Subject,
		HtmlBody:      email.HTML,
		TextBody:      email.Text,
		MessageStream: "outbound",
	}

	var resp sendResponse
	headers := map[string]string{authHeader: c.cfg.APIKey}

	if err := c.base.DoJSON(
		ctx,
		http.MethodPost,
		c.base.URL("/email"),
		req, &resp, headers,
	); err != nil {
		return err
	}

	if resp.ErrorCode != 0 {
		return fmt.Errorf(
			"postmark: send rejected (code %d): %s",
			resp.ErrorCode, resp.Message,
		)
	}

	return nil

}
