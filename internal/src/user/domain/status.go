package domain

type UserStatus string

const (
	StatusPending   UserStatus = "pending"
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusSuspended, StatusDeleted:
		return true
	default:
		return false
	}
}
