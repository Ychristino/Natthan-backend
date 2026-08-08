package address

import "time"

// ─── Requests ─────────────────────────────────────────────────────────────────

type CreateRequest struct {
	PersonID     string `json:"person_id"    validate:"required,uuid"`
	AddressType  string `json:"address_type" validate:"required,oneof=comercial residencial"`
	Street       string `json:"street"       validate:"required,min=1,max=255"`
	Number       string `json:"number"       validate:"required,max=20"`
	Complement   string `json:"complement"   validate:"max=100"`
	Neighborhood string `json:"neighborhood" validate:"required,max=100"`
	ZipCode      string `json:"zip_code"     validate:"required,max=10"`
	CityID       string `json:"city_id"      validate:"required,uuid"`
}

type UpdateRequest struct {
	AddressType  *string `json:"address_type" validate:"omitempty,oneof=comercial residencial"`
	Street       *string `json:"street"       validate:"omitempty,min=1,max=255"`
	Number       *string `json:"number"       validate:"omitempty,max=20"`
	Complement   *string `json:"complement"   validate:"omitempty,max=100"`
	Neighborhood *string `json:"neighborhood" validate:"omitempty,max=100"`
	ZipCode      *string `json:"zip_code"     validate:"omitempty,max=10"`
	CityID       *string `json:"city_id"      validate:"omitempty,uuid"`
}

type ListRequest struct {
	PersonID string `query:"person_id" validate:"omitempty,uuid"`
	Page     int    `query:"page"      validate:"min=1"`
	Limit    int    `query:"limit"     validate:"min=1,max=100"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type Response struct {
	ID           string       `json:"id"`
	PersonID     string       `json:"person_id"`
	AddressType  string       `json:"address_type"`
	Street       string       `json:"street"`
	Number       string       `json:"number"`
	Complement   string       `json:"complement,omitempty"`
	Neighborhood string       `json:"neighborhood"`
	ZipCode      string       `json:"zip_code"`
	City         CityResponse `json:"city"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
}

type ListResponse struct {
	Data  []Response `json:"data"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

type CityResponse struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	State StateResponse `json:"state"`
}

type StateResponse struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Code    string          `json:"code"`
	Country CountryResponse `json:"country"`
}

type CountryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// ToResponse converts an AddressDetail (domain model with JOIN data) to its HTTP response shape.
func ToResponse(a AddressDetail) Response {
	return Response{
		ID:           a.ID.String(),
		PersonID:     a.PersonID.String(),
		AddressType:  a.AddressType,
		Street:       a.Street,
		Number:       a.Number,
		Complement:   a.Complement,
		Neighborhood: a.Neighborhood,
		ZipCode:      a.ZipCode,
		City: CityResponse{
			ID:   a.CityID.String(),
			Name: a.CityName,
			State: StateResponse{
				ID:   a.StateID.String(),
				Name: a.StateName,
				Code: a.StateCode,
				Country: CountryResponse{
					ID:   a.CountryID.String(),
					Name: a.CountryName,
					Code: a.CountryCode,
				},
			},
		},
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
}
