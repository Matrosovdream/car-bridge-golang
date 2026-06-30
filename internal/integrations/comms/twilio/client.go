// Package twilio implements comms.SMSSender using the Twilio REST API.
//
// Twilio's Messages endpoint uses HTTP Basic auth and a form-encoded body
// rather than the JSON/Bearer convention of the shared client, so this client
// issues the request itself instead of going through Client.DoJSON.
package twilio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/comms"
)

const maxBodyBytes = 1 << 20 // 1 MiB cap on response bodies

// Client sends SMS through the Twilio REST API.
type Client struct {
	base       *integrations.Client
	cfg        integrations.HTTPConfig
	http       *http.Client
	accountSID string
}

// New builds a twilio client from config and a shared base HTTP client. The
// Account SID is read from cfg.BaseURL (.../Accounts/<SID>) and reused as the
// Basic-auth username; cfg.APIKey is the auth token.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		base:       base,
		cfg:        cfg,
		http:       &http.Client{Timeout: timeout},
		accountSID: accountSIDFromURL(cfg.BaseURL),
	}
}

// compile-time check that Client satisfies the port.
var _ comms.SMSSender = (*Client)(nil)

// Send delivers a single SMS via Twilio's Messages endpoint.
func (c *Client) Send(ctx context.Context, msg comms.Message) error {

	if c.accountSID == "" {
		return fmt.Errorf("twilio: account SID missing from base_url")
	}
	if msg.From == "" {
		return fmt.Errorf("twilio: message.From is required")
	}

	form := url.Values{}
	form.Set("To", msg.To)
	form.Set("From", msg.From)
	form.Set("Body", msg.Body)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.base.URL("/Messages.json"),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("twilio: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.accountSID, c.cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", integrations.ErrUpstreamUnavailable, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return integrations.ErrRateLimited
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: status %d", integrations.ErrUpstreamUnavailable, resp.StatusCode)
	default:
		// Twilio returns {"code":..., "message":...} on errors.
		var e struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(data, &e)
		if e.Message != "" {
			return fmt.Errorf("twilio: send failed (status %d, code %d): %s", resp.StatusCode, e.Code, e.Message)
		}
		return fmt.Errorf("twilio: unexpected status %d", resp.StatusCode)
	}

}

// accountSIDFromURL extracts the Account SID from a Twilio REST base URL of the
// form https://api.twilio.com/2010-04-01/Accounts/<SID>.
func accountSIDFromURL(base string) string {
	parts := strings.Split(strings.Trim(base, "/"), "/")
	for i, p := range parts {
		if p == "Accounts" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
