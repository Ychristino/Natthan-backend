package contact

import "time"

// ─── Requests ─────────────────────────────────────────────────────────────────

type CreateRequest struct {
	PersonID string `json:"person_id" validate:"required,uuid"`
	Type     string `json:"type"      validate:"required,oneof=email phone whatsapp"`
	Value    string `json:"value"     validate:"required,min=1,max=255"`
}

type ListRequest struct {
	PersonID string `query:"person_id" validate:"omitempty,uuid"`
	Page     int    `query:"page"      validate:"min=1"`
	Limit    int    `query:"limit"     validate:"min=1,max=100"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type Response struct {
	ID        string `json:"id"`
	PersonID  string `json:"person_id"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

type ListResponse struct {
	Data  []Response `json:"data"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func ToResponse(c Contact) Response {
	return Response{
		ID:        c.ID.String(),
		PersonID:  c.PersonID.String(),
		Type:      c.Type,
		Value:     c.Value,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}
