package service

import "time"

// ─── Requests ─────────────────────────────────────────────────────────────────

type CreateRequest struct {
	ServiceCode string  `json:"service_code" validate:"required,max=100"`
	Name        string  `json:"name"         validate:"required,min=2,max=255"`
	Description string  `json:"description"  validate:"max=1000"`
	Price       float64 `json:"price"        validate:"gte=0"`
}

type UpdateRequest struct {
	ServiceCode *string  `json:"service_code" validate:"omitempty,max=100"`
	Name        *string  `json:"name"         validate:"omitempty,min=2,max=255"`
	Description *string  `json:"description"  validate:"omitempty,max=1000"`
	Price       *float64 `json:"price"        validate:"omitempty,gte=0"`
}

type ListRequest struct {
	Name  string `query:"name"`
	Page  int    `query:"page"  validate:"min=1"`
	Limit int    `query:"limit" validate:"min=1,max=100"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type Response struct {
	ID          string  `json:"id"`
	ServiceCode string  `json:"service_code"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Active      bool    `json:"active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ListResponse struct {
	Data  []Response `json:"data"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toResponse(s Service) Response {
	return Response{
		ID:          s.ID.String(),
		ServiceCode: s.ServiceCode,
		Name:        s.Name,
		Description: s.Description,
		Price:       s.Price,
		Active:      s.Active,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}
