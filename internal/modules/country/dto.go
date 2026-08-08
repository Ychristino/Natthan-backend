package country

import "time"

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Code string `json:"code" validate:"required,len=2"`
}

type UpdateRequest struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
	Code *string `json:"code" validate:"omitempty,len=2"`
}

type ListRequest struct {
	Name  string `query:"name"`
	Page  int    `query:"page"  validate:"min=1"`
	Limit int    `query:"limit" validate:"min=1,max=300"`
}

type Response struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListResponse struct {
	Data  []Response `json:"data"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toResponse(c Country) Response {
	return Response{
		ID:        c.ID.String(),
		Name:      c.Name,
		Code:      c.Code,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
}
