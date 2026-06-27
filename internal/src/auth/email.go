package auth

import "context"

type EmailSender interface {
	SendVerification(ctx context.Context, to, verificationURL string) error

	SendPasswordReset(ctx context.Context, to, resetURL string) error
}
