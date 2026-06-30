// Package comms defines notification capability ports.
// Providers (twilio, postmark, sendgrid) implement these.
package comms

import "context"

// Message is an outbound SMS.
type Message struct {
	From string
	To   string
	Body string
}

// Email is an outbound email.
type Email struct {
	From    string
	To      string
	Subject string
	HTML    string
	Text    string
}

// SMSSender sends text messages (e.g. Twilio).
type SMSSender interface {
	Send(ctx context.Context, msg Message) error
}

// EmailSender sends email (e.g. Postmark, SendGrid).
type EmailSender interface {
	Send(ctx context.Context, email Email) error
}
