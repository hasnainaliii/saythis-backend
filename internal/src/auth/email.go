// Package email provides an abstraction for sending transactional emails.
// The EmailSender interface decouples the application from any specific provider,
// making it easy to swap implementations (Resend, SendGrid, SMTP, etc.) or stub
// the sender in tests.
package auth

import "context"

// EmailSender is the interface every email provider must implement.
type EmailSender interface {
	// SendVerification sends the one-time email verification link to a new user.
	SendVerification(ctx context.Context, to, verificationURL string) error

	// SendPasswordReset sends a password-reset link to a user who requested it.
	SendPasswordReset(ctx context.Context, to, resetURL string) error
}
