package auth

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
	RoleVisitor  Role = "visitor"
)

type UserRole struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Role           Role
	ExpirationDate *time.Time
	Active         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type User struct {
	ID           uuid.UUID
	PersonID     uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	Active       bool
	Roles        []UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// effectiveRole returns the most permissive active non-expired role.
func effectiveRole(roles []UserRole) Role {
	now := time.Now()
	check := func(target Role) bool {
		for _, r := range roles {
			if r.Role != target || !r.Active {
				continue
			}
			if r.ExpirationDate != nil && r.ExpirationDate.Before(now) {
				continue
			}
			return true
		}
		return false
	}
	if check(RoleAdmin) {
		return RoleAdmin
	}
	if check(RoleEmployee) {
		return RoleEmployee
	}
	return RoleVisitor
}
