package city

import "time"

type CreateRequest struct {
	Name    string `json:"name"     validate:"required,min=2,max=100"`
	StateID string `json:"state_id" validate:"required,uuid"`
}

type UpdateRequest struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type ListRequest struct {
	StateID string `query:"state_id" validate:"omitempty,uuid"`
	Name    string `query:"name"`
	Page    int    `query:"page"  validate:"min=1"`
	Limit   int    `query:"limit" validate:"min=1,max=1000"`
}

type Response struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StateID   string `json:"state_id"`
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

func toResponse(c City) Response {
	return Response{
		ID:        c.ID.String(),
		Name:      c.Name,
		StateID:   c.StateID.String(),
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
}
