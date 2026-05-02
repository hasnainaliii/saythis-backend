package domain

type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleAdmin     UserRole = "admin"
	RoleTherapist UserRole = "therapist"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleUser, RoleAdmin, RoleTherapist:
		return true
	default:
		return false
	}
}
