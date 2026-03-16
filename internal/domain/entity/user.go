package entity

// User roles
const (
	UserRoleUser        = "USER"
	UserRoleAdmin       = "ADMIN"
	UserRoleSystemAdmin = "SYSTEM_ADMIN"
)

// User statuses
const (
	UserStatusActive   = "ACTIVE"
	UserStatusInactive = "INACTIVE"
)

// IsValidUserRole checks if the role is valid
func IsValidUserRole(role string) bool {
	switch role {
	case UserRoleUser, UserRoleAdmin, UserRoleSystemAdmin:
		return true
	default:
		return false
	}
}

// IsValidUserStatus checks if the status is valid
func IsValidUserStatus(status string) bool {
	switch status {
	case UserStatusActive, UserStatusInactive:
		return true
	default:
		return false
	}
}
