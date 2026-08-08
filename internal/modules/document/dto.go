package document

import "time"

// ─── Requests ─────────────────────────────────────────────────────────────────

type CreateRequest struct {
	PersonID string `json:"person_id" validate:"required,uuid"`
	Type     string `json:"type"      validate:"required,oneof=cpf cnpj rg cnh passport state_registration"`
	Value    string `json:"value"     validate:"required,min=1,max=50"`
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

func ToResponse(d Document) Response {
	return Response{
		ID:        d.ID.String(),
		PersonID:  d.PersonID.String(),
		Type:      d.Type,
		Value:     d.Value,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
	}
}
