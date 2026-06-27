package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const resendAPIURL = "https://api.resend.com/emails"

type ResendClient struct {
	apiKey     string
	from       string
	httpClient *http.Client
}

func NewResendClient(apiKey, from string) *ResendClient {
	return &ResendClient{
		apiKey: apiKey,
		from:   from,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (c *ResendClient) send(ctx context.Context, to, subject, html string) error {
	payload := resendRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("resend: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend: unexpected status %d", resp.StatusCode)
	}

	slog.Debug("email sent via resend", "to", to, "subject", subject)
	return nil
}

func (c *ResendClient) SendVerification(ctx context.Context, to, verificationURL string) error {
	return c.send(ctx, to, "Verify your email address", buildVerificationHTML(verificationURL))
}

func (c *ResendClient) SendPasswordReset(ctx context.Context, to, resetURL string) error {
	return c.send(ctx, to, "Reset your password", buildPasswordResetHTML(resetURL))
}
