package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

const (
	minFullNameLength = 3
	maxFullNameLength = 100
)

type User struct {
	id              uuid.UUID
	email           string
	fullName        string
	avatarURL       string
	role            UserRole
	status          UserStatus
	emailVerifiedAt *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func NewUser(email string, fullName string, role UserRole, timeNow time.Time) (*User, error) {

	email = strings.ToLower(strings.TrimSpace(email))
	fullName = strings.TrimSpace(fullName)

	if email == "" {
		return nil, ErrEmptyEmail
	}
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if fullName == "" {
		return nil, ErrEmptyFullName
	}
	if len(fullName) < minFullNameLength || len(fullName) > maxFullNameLength {
		return nil, ErrInvalidFullNameLength
	}
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	return &User{
		id:        uuid.New(),
		email:     email,
		fullName:  fullName,
		role:      role,
		status:    StatusActive,
		createdAt: timeNow,
		updatedAt: timeNow,
	}, nil
}

// *******
// Getters
// *******

func (u *User) ID() uuid.UUID               { return u.id }
func (u *User) Email() string               { return u.email }
func (u *User) FullName() string            { return u.fullName }
func (u *User) AvatarURL() string           { return u.avatarURL }
func (u *User) Role() UserRole              { return u.role }
func (u *User) Status() UserStatus          { return u.status }
func (u *User) CreatedAt() time.Time        { return u.createdAt }
func (u *User) UpdatedAt() time.Time        { return u.updatedAt }
func (u *User) EmailVerifiedAt() *time.Time { return u.emailVerifiedAt }

// *******
// Setters
// *******

func (u *User) SetCreatedAt(createdAt time.Time) { u.createdAt = createdAt }
func (u *User) SetUpdatedAt(updatedAt time.Time) { u.updatedAt = updatedAt }
