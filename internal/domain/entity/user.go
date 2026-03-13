package entity

import (
	"time"
)

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

// User represents the user entity
type User struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

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
