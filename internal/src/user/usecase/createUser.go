package usecase

import (
	"context"
	"saythis-backend/internal/src/user/domain"
	"saythis-backend/internal/src/user/repository"
	"time"

	"go.uber.org/zap"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) WithRepository(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		repo: repo,
	}
}

func (uc *UserUseCase) CreateUser(ctx context.Context, email, fullName string) (*domain.User, error) {

	role := domain.RoleUser
	timeNow := time.Now()

	user, err := domain.NewUser(email, fullName, role, timeNow)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		zap.S().Errorw("Failed to create user", "email", email, "error", err)
		return nil, err
	}

	zap.S().Debugw("User created successfully", "email", email, "user_id", user.ID())
	return user, nil
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, userID string) error {
	if err := uc.repo.SoftDelete(ctx, userID); err != nil {
		zap.S().Errorw("Failed to delete user", "user_id", userID, "error", err)
		return err
	}

	zap.S().Infow("User deleted successfully", "user_id", userID)
	return nil
}

// UpdateProfileInput represents optional fields for profile updates
type UpdateProfileInput struct {
	FullName  *string
	AvatarURL *string
}

// UpdateProfile updates user profile fields based on provided input
func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) error {
	updates := make(map[string]interface{})

	if input.FullName != nil {
		// Validate full name length
		name := *input.FullName
		if len(name) < 2 || len(name) > 255 {
			return &ProfileValidationError{Field: "full_name", Message: "name must be between 2 and 255 characters"}
		}
		updates["full_name"] = name
	}

	if input.AvatarURL != nil {
		updates["avatar_url"] = *input.AvatarURL
	}

	if len(updates) == 0 {
		return &ProfileValidationError{Field: "", Message: "at least one field must be provided"}
	}

	if err := uc.repo.Update(ctx, userID, updates); err != nil {
		zap.S().Errorw("Failed to update user profile", "user_id", userID, "error", err)
		return err
	}

	zap.S().Infow("User profile updated successfully", "user_id", userID)
	return nil
}

// ProfileValidationError represents a validation error for profile updates
type ProfileValidationError struct {
	Field   string
	Message string
}

func (e *ProfileValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}
