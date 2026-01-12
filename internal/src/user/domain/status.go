package domain

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusSuspended, StatusDeleted:
		return true
	default:
		return false
	}
}
