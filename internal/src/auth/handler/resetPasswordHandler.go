package handler

import (
	"encoding/json"
	"net/http"
	"saythis-backend/internal/apperror"
	"saythis-backend/internal/src/auth/usecase"

	"go.uber.org/zap"
)

type ResetPasswordHandler struct {
	resetPasswordUC *usecase.ResetPasswordUseCase
}

func NewResetPasswordHandler(resetPasswordUC *usecase.ResetPasswordUseCase) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		resetPasswordUC: resetPasswordUC,
	}
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	// Handle GET - show a simple form for testing
	if r.Method == http.MethodGet {
		if token == "" {
			h.respondWithError(w, "MISSING_TOKEN", "Token is required in URL", http.StatusBadRequest)
			return
		}

		// Show a simple HTML form that POSTs directly
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Reset Password</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
        button { width: 100%; padding: 12px; background: #667eea; color: white; border: none; border-radius: 5px; cursor: pointer; }
    </style>
</head>
<body>
    <h2>Reset Your Password</h2>
    <form method="POST" action="/api/v1/auth/reset-password?token=` + token + `">
        <input type="password" name="new_password" placeholder="New Password (min 8 chars)" required minlength="8">
        <button type="submit">Reset Password</button>
    </form>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		return
	}

	if r.Method != http.MethodPost {
		h.respondWithError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqToken, reqPassword string

	// Check Content-Type to handle both form and JSON
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		// Handle form POST
		r.ParseForm()
		reqToken = r.URL.Query().Get("token") // Token from URL
		reqPassword = r.FormValue("new_password")
	} else {
		// Handle JSON POST
		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, "INVALID_REQUEST", "Invalid request body", http.StatusBadRequest)
			return
		}
		reqToken = req.Token
		reqPassword = req.NewPassword
	}

	if reqToken == "" {
		h.respondWithError(w, "VALIDATION_ERROR", "Token is required", http.StatusBadRequest)
		return
	}

	if reqPassword == "" {
		h.respondWithError(w, "VALIDATION_ERROR", "New password is required", http.StatusBadRequest)
		return
	}

	input := usecase.ResetPasswordInput{
		Token:       reqToken,
		NewPassword: reqPassword,
	}

	ctx := r.Context()
	if err := h.resetPasswordUC.Execute(ctx, input); err != nil {
		h.handleError(w, err)
		return
	}

	// For form submissions, show HTML response
	if contentType == "application/x-www-form-urlencoded" {
		html := `<!DOCTYPE html>
<html>
<head><title>Password Reset</title>
<style>body{font-family:Arial;max-width:400px;margin:50px auto;padding:20px;text-align:center;}</style>
</head>
<body>
<h2 style="color:green;">✓ Password Reset Successfully!</h2>
<p>You can now login with your new password.</p>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		return
	}

	// For JSON requests
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ResetPasswordResponse{
		Message: "Password has been reset successfully",
	})
}

func (h *ResetPasswordHandler) handleError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		h.respondWithError(w, appErr.Code, appErr.Message, appErr.HTTPStatus)
		return
	}

	zap.S().Errorw("Reset password failed", "error", err)
	h.respondWithError(w, "INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
}

func (h *ResetPasswordHandler) respondWithError(w http.ResponseWriter, code, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
