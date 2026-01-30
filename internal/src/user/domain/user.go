package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	id              uuid.UUID
	email           string
	fullName        string
	avatarURL       *string
	role            UserRole
	status          UserStatus
	emailVerifiedAt *time.Time
	createdAt       time.Time
	updatedAt       time.Time
	deletedAt       *time.Time
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

const (
	minFullNameLength = 2
	maxFullNameLength = 255
)

func NewUser(email string, fullName string, role UserRole, TimeNow time.Time) (*User, error) {

	id := uuid.New()
	if len(fullName) < minFullNameLength || len(fullName) > maxFullNameLength {
		return nil, ErrInvalidFullNameLength
	}
	if strings.TrimSpace(email) == "" {
		return nil, ErrEmptyEmail
	}

	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	if strings.TrimSpace(fullName) == "" {
		return nil, ErrEmptyFullName
	}

	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	return &User{
		id:        id,
		email:     strings.ToLower(email),
		fullName:  fullName,
		role:      role,
		status:    StatusActive,
		createdAt: TimeNow,
		updatedAt: TimeNow,
	}, nil
}

func (u *User) ID() uuid.UUID               { return u.id }
func (u *User) Email() string               { return u.email }
func (u *User) FullName() string            { return u.fullName }
func (u *User) Role() UserRole              { return u.role }
func (u *User) Status() UserStatus          { return u.status }
func (u *User) CreatedAt() time.Time        { return u.createdAt }
func (u *User) UpdatedAt() time.Time        { return u.updatedAt }
func (u *User) EmailVerifiedAt() *time.Time { return u.emailVerifiedAt }

func (u *User) SetCreatedAt(t time.Time) {
	u.createdAt = t
}

func (u *User) SetUpdatedAt(t time.Time) {
	u.updatedAt = t
}

func (u *User) SetEmailVerifiedAt(t *time.Time) {
	u.emailVerifiedAt = t
}

func (u *User) ChangeFullName(name string, now time.Time) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyFullName
	}
	u.fullName = name
	u.updatedAt = now
	return nil
}

func (u *User) VerifyEmail(now time.Time) {
	u.emailVerifiedAt = &now
	u.updatedAt = now
}

func (u *User) Suspend(now time.Time) error {
	if u.status == StatusDeleted {
		return ErrInvalidStatus
	}
	u.status = StatusSuspended
	u.updatedAt = now
	return nil
}

func (u *User) DeletedAt() *time.Time {
	return u.deletedAt
}

func (u *User) MarkAsDeleted(now time.Time) {
	u.deletedAt = &now
	u.status = StatusDeleted
	u.updatedAt = now
}
