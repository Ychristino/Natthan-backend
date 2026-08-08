package person

import (
	"time"

	"github.com/natthan/api/internal/modules/address"
	"github.com/natthan/api/internal/modules/contact"
	"github.com/natthan/api/internal/modules/document"
)

// ─── Requests ─────────────────────────────────────────────────────────────────

type DocumentInlineRequest struct {
	Type  string `json:"type"  validate:"required,oneof=cpf cnpj rg cnh passport state_registration"`
	Value string `json:"value" validate:"required,min=1,max=50"`
}

type AddressInlineRequest struct {
	AddressType  string `json:"address_type" validate:"required,oneof=comercial residencial"`
	Street       string `json:"street"       validate:"required,min=1,max=255"`
	Number       string `json:"number"       validate:"required,max=20"`
	Complement   string `json:"complement"   validate:"max=100"`
	Neighborhood string `json:"neighborhood" validate:"required,max=100"`
	ZipCode      string `json:"zip_code"     validate:"required,max=10"`
	CityID       string `json:"city_id"      validate:"required,uuid"`
}

type ContactInlineRequest struct {
	Type  string `json:"type"  validate:"required,oneof=email phone whatsapp"`
	Value string `json:"value" validate:"required,min=1,max=255"`
}

type CreateRequest struct {
	FullName      string                  `json:"full_name"      validate:"required,min=2,max=255"`
	BirthDate     *string                `json:"birth_date"     validate:"omitempty,datetime=2006-01-02"`
	NationalityID *string                `json:"nationality_id" validate:"omitempty,uuid"`
	Documents     []DocumentInlineRequest `json:"documents"`
	Addresses     []AddressInlineRequest  `json:"addresses"`
	Contacts      []ContactInlineRequest  `json:"contacts"`
}

type UpdateRequest struct {
	FullName      *string `json:"full_name"      validate:"omitempty,min=2,max=255"`
	BirthDate     *string `json:"birth_date"     validate:"omitempty,datetime=2006-01-02"`
	NationalityID *string `json:"nationality_id" validate:"omitempty,uuid"`
}

type ListRequest struct {
	Name  string `query:"name"`
	Page  int    `query:"page"  validate:"min=1"`
	Limit int    `query:"limit" validate:"min=1,max=100"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type NationalityResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type PersonResponse struct {
	ID          string               `json:"id"`
	FullName    string               `json:"full_name"`
	BirthDate   *string             `json:"birth_date,omitempty"`
	Nationality *NationalityResponse `json:"nationality,omitempty"`
	Documents   []document.Response  `json:"documents,omitempty"`
	Addresses   []address.Response   `json:"addresses,omitempty"`
	Contacts    []contact.Response   `json:"contacts,omitempty"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

type ListResponse struct {
	Data  []PersonResponse `json:"data"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

type LookupResponse struct {
	Found    bool   `json:"found"`
	PersonID string `json:"person_id,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toResponse(p Person) PersonResponse {
	resp := PersonResponse{
		ID:        p.ID.String(),
		FullName:  p.FullName,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
	if p.BirthDate != nil {
		s := p.BirthDate.Format("2006-01-02")
		resp.BirthDate = &s
	}
	if p.NationalityID != nil {
		resp.Nationality = &NationalityResponse{
			ID:   p.NationalityID.String(),
			Name: p.NationalityName,
			Code: p.NationalityCode,
		}
	}
	return resp
}

func toDocumentResponses(docs []document.Document) []document.Response {
	if len(docs) == 0 {
		return nil
	}
	resp := make([]document.Response, len(docs))
	for i, d := range docs {
		resp[i] = document.ToResponse(d)
	}
	return resp
}

func toAddressResponses(addrs []address.AddressDetail) []address.Response {
	if len(addrs) == 0 {
		return nil
	}
	resp := make([]address.Response, len(addrs))
	for i, a := range addrs {
		resp[i] = address.ToResponse(a)
	}
	return resp
}

func toContactResponses(contacts []contact.Contact) []contact.Response {
	if len(contacts) == 0 {
		return nil
	}
	resp := make([]contact.Response, len(contacts))
	for i, c := range contacts {
		resp[i] = contact.ToResponse(c)
	}
	return resp
}
