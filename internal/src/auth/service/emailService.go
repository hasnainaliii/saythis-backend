package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type EmailService struct {
	apiKey  string
	baseURL string
}

func NewEmailService(apiKey string, appBaseURL string) *EmailService {
	return &EmailService{
		apiKey:  apiKey,
		baseURL: appBaseURL,
	}
}

type ResendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type ResendResponse struct {
	ID string `json:"id"`
}

func (s *EmailService) SendPasswordResetEmail(email, token string) error {
	resetURL := fmt.Sprintf("%s/api/v1/auth/reset-password?token=%s", s.baseURL, token)

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">Password Reset Request</h1>
    </div>
    <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hello,</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold;">Reset Password</a>
        </div>
        <p>Or copy and paste this link into your browser:</p>
        <p style="background: #eee; padding: 10px; border-radius: 5px; word-break: break-all;">%s</p>
        <p><strong>This link will expire in 20 minutes.</strong></p>
        <p>If you didn't request this, please ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">
        <p style="color: #666; font-size: 12px;">This is an automated message, please do not reply.</p>
    </div>
</body>
</html>
`, resetURL, resetURL)

	reqBody := ResendRequest{
		From:    "SayThis <onboarding@resend.dev>",
		To:      []string{email},
		Subject: "Reset Your Password - SayThis",
		HTML:    htmlContent,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	apiKeyPreview := "EMPTY"
	if len(s.apiKey) > 10 {
		apiKeyPreview = s.apiKey[:10] + "..."
	}
	zap.S().Infow("Sending email via Resend", "to", email, "apiKeyPrefix", apiKeyPreview)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Errorw("HTTP request to Resend failed", "error", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	zap.S().Infow("Resend API response", "statusCode", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		zap.S().Errorw("Resend API error", "status", resp.StatusCode, "body", errorBody, "recipient", email)
		return fmt.Errorf("resend API returned status %d: %v", resp.StatusCode, errorBody)
	}

	var resendResp ResendResponse
	if err := json.NewDecoder(resp.Body).Decode(&resendResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	zap.S().Infow("Password reset email sent", "email_id", resendResp.ID, "recipient", email)
	return nil
}
