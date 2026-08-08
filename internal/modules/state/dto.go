package state

import "time"

type CreateRequest struct {
	Name      string `json:"name"       validate:"required,min=2,max=100"`
	Code      string `json:"code"       validate:"required,len=2"`
	CountryID string `json:"country_id" validate:"required,uuid"`
}

type UpdateRequest struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
	Code *string `json:"code" validate:"omitempty,len=2"`
}

type ListRequest struct {
	CountryID string `query:"country_id" validate:"omitempty,uuid"`
	Name      string `query:"name"`
	Page      int    `query:"page"  validate:"min=1"`
	Limit     int    `query:"limit" validate:"min=1,max=100"`
}

type Response struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	CountryID string `json:"country_id"`
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

func toResponse(st State) Response {
	return Response{
		ID:        st.ID.String(),
		Name:      st.Name,
		Code:      st.Code,
		CountryID: st.CountryID.String(),
		CreatedAt: st.CreatedAt.Format(time.RFC3339),
		UpdatedAt: st.UpdatedAt.Format(time.RFC3339),
	}
}
