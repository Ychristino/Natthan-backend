package auth

import "time"

// ─── Auth Requests ────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Username    string  `json:"username"    validate:"required,min=3,max=50"`
	Email       string  `json:"email"       validate:"required,email"`
	Password    string  `json:"password"    validate:"required,min=6"`
	FullName      string  `json:"full_name"      validate:"required,min=2,max=255"`
	BirthDate     *string `json:"birth_date"     validate:"omitempty,datetime=2006-01-02"`
	NationalityID *string `json:"nationality_id" validate:"omitempty,uuid"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ─── User Management Requests ─────────────────────────────────────────────────

type AddRoleRequest struct {
	Role           string  `json:"role"            validate:"required,oneof=admin employee visitor"`
	ExpirationDate *string `json:"expiration_date" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type ListUsersRequest struct {
	Name   string `query:"name"`
	Role   string `query:"role"   validate:"omitempty,oneof=admin employee visitor"`
	Active string `query:"active"` // "true" (default) | "false"
	Page   int    `query:"page"   validate:"min=1"`
	Limit  int    `query:"limit"  validate:"min=1,max=100"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserRoleResponse struct {
	Role           string  `json:"role"`
	ExpirationDate *string `json:"expiration_date,omitempty"`
	Active         bool    `json:"active"`
}

type UserResponse struct {
	ID        string             `json:"id"`
	PersonID  string             `json:"person_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Active    bool               `json:"active"`
	Roles     []UserRoleResponse `json:"roles"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

type UserListResponse struct {
	Data  []UserResponse `json:"data"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toUserResponse(u *User) UserResponse {
	roles := make([]UserRoleResponse, len(u.Roles))
	for i, r := range u.Roles {
		ur := UserRoleResponse{Role: string(r.Role), Active: r.Active}
		if r.ExpirationDate != nil {
			s := r.ExpirationDate.Format(time.RFC3339)
			ur.ExpirationDate = &s
		}
		roles[i] = ur
	}
	return UserResponse{
		ID:        u.ID.String(),
		PersonID:  u.PersonID.String(),
		Username:  u.Username,
		Email:     u.Email,
		Active:    u.Active,
		Roles:     roles,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func toUserRoleResponse(ur *UserRole) UserRoleResponse {
	r := UserRoleResponse{Role: string(ur.Role), Active: ur.Active}
	if ur.ExpirationDate != nil {
		s := ur.ExpirationDate.Format(time.RFC3339)
		r.ExpirationDate = &s
	}
	return r
}
