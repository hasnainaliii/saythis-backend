package handler

import (
	"net/http"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/user/usecase"
)

// allowedImageTypes is the set of MIME types accepted for avatar uploads.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// UpdateAvatarHandler handles PATCH /api/v1/users/me/avatar.
// Expects a multipart/form-data request with a single "avatar" file field (max 5 MB).
// The image is uploaded to Cloudinary and the returned secure URL is persisted.
type UpdateAvatarHandler struct {
	usecase *usecase.UserUseCase
}

func NewUpdateAvatarHandler(uc *usecase.UserUseCase) *UpdateAvatarHandler {
	return &UpdateAvatarHandler{usecase: uc}
}

type updateAvatarResponse struct {
	User userPayload `json:"user"`
}

func (h *UpdateAvatarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxSize = 5 << 20 // 5 MB

	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := r.ParseMultipartForm(maxSize); err != nil {
		helper.Error(w, http.StatusBadRequest, "file exceeds 5 MB limit or invalid multipart form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("avatar")
	if err != nil {
		helper.Error(w, http.StatusBadRequest, "avatar field is required")
		return
	}
	defer file.Close()

	// Validate content type declared in the multipart header.
	ct := header.Header.Get("Content-Type")
	if !allowedImageTypes[ct] {
		helper.Error(w, http.StatusBadRequest, "only JPEG, PNG, WebP, and GIF images are allowed")
		return
	}

	user, err := h.usecase.UpdateAvatar(r.Context(), claims.UserID, file, header.Filename)
	if err != nil {
		status, msg := mapUserError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, updateAvatarResponse{
		User: userPayload{
			ID:        user.ID(),
			Email:     user.Email(),
			FullName:  user.FullName(),
			AvatarURL: user.AvatarURL(),
			Role:      user.Role(),
			Status:    user.Status(),
			CreatedAt: user.CreatedAt(),
			UpdatedAt: user.UpdatedAt(),
		},
	})
}
