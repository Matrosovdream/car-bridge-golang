package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxBodyBytes = 1 << 20
	baseBackoff = 200 * time.Millisecond
)

type Client struct {
	cfg HTTPConfig
	http *http.Client
	log *logrus.Logger
}

func NewClient(
	cfg HTTPConfig, log *logrus.Logger,
) *Client {

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: timeout,
		},
		log: log,
	}

}

func (c *Client) Config() HTTPConfig {
	return c.cfg
}

func (c *Client) URL(path string) string {

	return strings.TrimRight(c.cfg.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")

}

func (c *Client) DoJSON(
	ctx context.Context,
	method, url string,
	body, out any,
	headers map[string]string,
) error {

	retries := c.cfg.Retries
	if retries < 0 {
		retries = 0
	}

	var lastErr error 
	for attempt := 0; attempt <= retries; attempt++ {

		if attempt > 0 {
			wait := time.Duration(math.Pow(2, float64(attempt-1))) * baseBackoff

			select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(wait):
			}
		}

		err := c.do(
			ctx, method, url, body, out, headers,
		)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryable(err) {
			return err
		}
		if c.log != nil {
			c.log.Warnf(
				"integration %s %s failed (attempt %d/%d): %v",
				method, url, attempt+1, retries+1, err,
			)
		}

	}

	return lastErr

}

func (c *Client) do(
	ctx context.Context,
	method, url string,
	body, out any,
	headers map[string]string,
) error {

	var reqBody io.Reader

	if body != nil {
		
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf(
				"marshal request body: %w", err,
			)
		}
		reqBody = bytes.NewReader(b)

	}

	req, err := http.NewRequestWithContext(
		ctx, method, url, reqBody,
	)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf(
			"%w: %v",
			ErrUpstreamUnavailable, err,
		)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(
		io.LimitReader(resp.Body, maxBodyBytes),
	)

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if out == nil || len(data) == 0 {
			return nil
		}
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("%w: %v", ErrBadResponse, err)
		}
		return nil
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", ErrRateLimited, truncate(data))
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: status %d: %s", ErrUpstreamUnavailable, resp.StatusCode, truncate(data))
	default:
		return fmt.Errorf("integration: unexpected status %d: %s", resp.StatusCode, truncate(data))
	}

}

func isRetryable(err error) bool {
	return errors.Is(err, ErrUpstreamUnavailable) ||
		errors.Is(err, ErrRateLimited)
}

func truncate(b []byte) string {

	const max = 256
	if len(b) > max {
		return string(b[:max]) + "..."
	}

	return string(b)

}