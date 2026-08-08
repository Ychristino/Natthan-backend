package serviceOrder

import "time"

// ─── Requests ─────────────────────────────────────────────────────────────────

type CreateRequest struct {
	ClientID string              `json:"client_id" validate:"required,uuid"`
	Status   string              `json:"status"    validate:"omitempty,oneof=open in_progress completed cancelled"`
	Notes    string              `json:"notes"     validate:"max=1000"`
	PayDate  *string             `json:"pay_date"  validate:"omitempty"`
	IsPaid   bool                `json:"is_paid"`
	Products []AddProductRequest `json:"products"`
	Services []AddServiceRequest `json:"services"`
}

type UpdateRequest struct {
	Status   *string `json:"status"    validate:"omitempty,oneof=open in_progress completed cancelled"`
	ExitDate *string `json:"exit_date" validate:"omitempty"`
	PayDate  *string `json:"pay_date"  validate:"omitempty"`
	IsPaid   *bool   `json:"is_paid"   validate:"omitempty"`
	Notes    *string `json:"notes"     validate:"omitempty,max=1000"`
}

type ListRequest struct {
	Search   string `query:"search"`
	Status   string `query:"status"    validate:"omitempty,oneof=open in_progress completed cancelled"`
	DateFrom string `query:"date_from"`
	DateTo   string `query:"date_to"`
	Page     int    `query:"page"      validate:"min=1"`
	Limit    int    `query:"limit"     validate:"min=1,max=100"`
}

type AddProductRequest struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Quantity  int     `json:"quantity"   validate:"required,min=1"`
	UnitPrice float64 `json:"unit_price" validate:"gte=0"`
}

type AddServiceRequest struct {
	ServiceID    string  `json:"service_id"    validate:"required,uuid"`
	PriceApplied float64 `json:"price_applied" validate:"gte=0"`
	Notes        string  `json:"notes"         validate:"max=1000"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type OrderProductResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductCode string  `json:"product_code"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type OrderServiceResponse struct {
	ID           string  `json:"id"`
	ServiceID    string  `json:"service_id"`
	ServiceCode  string  `json:"service_code"`
	Name         string  `json:"name"`
	PriceApplied float64 `json:"price_applied"`
	Notes        string  `json:"notes,omitempty"`
}

// Response is returned by List — lean, no line items.
type Response struct {
	ID               string `json:"id"`
	ClientID         string `json:"client_id"`
	ClientName       string `json:"client_name"`
	EmployeeID       string `json:"employee_id"`
	EmployeeFullName string `json:"employee_full_name"`
	Status           string `json:"status"`
	EntryDate        string `json:"entry_date"`
	ExitDate         string `json:"exit_date,omitempty"`
	PayDate          string `json:"pay_date,omitempty"`
	IsPaid           bool   `json:"is_paid"`
	Notes            string `json:"notes,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// DetailResponse is returned by GetByID — includes all line items.
type DetailResponse struct {
	Response
	Products []OrderProductResponse `json:"products"`
	Services []OrderServiceResponse `json:"services"`
}

type ListResponse struct {
	Data  []Response `json:"data"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toResponse(so ServiceOrder) Response {
	r := Response{
		ID:               so.ID.String(),
		ClientID:         so.ClientID.String(),
		ClientName:       so.ClientName,
		EmployeeID:       so.EmployeeID.String(),
		EmployeeFullName: so.EmployeeFullName,
		Status:           string(so.Status),
		EntryDate:        so.EntryDate.Format(time.RFC3339),
		IsPaid:           so.IsPaid,
		Notes:            so.Notes,
		CreatedAt:        so.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        so.UpdatedAt.Format(time.RFC3339),
	}
	if so.ExitDate != nil {
		r.ExitDate = so.ExitDate.Format(time.RFC3339)
	}
	if so.PayDate != nil {
		r.PayDate = so.PayDate.Format(time.RFC3339)
	}
	return r
}

func toDetailResponse(so ServiceOrder) DetailResponse {
	products := make([]OrderProductResponse, len(so.Products))
	for i, p := range so.Products {
		products[i] = toOrderProductResponse(p)
	}
	services := make([]OrderServiceResponse, len(so.Services))
	for i, sv := range so.Services {
		services[i] = toOrderServiceResponse(sv)
	}
	return DetailResponse{
		Response: toResponse(so),
		Products: products,
		Services: services,
	}
}

func toOrderProductResponse(p OrderProduct) OrderProductResponse {
	return OrderProductResponse{
		ID:          p.ID.String(),
		ProductID:   p.ProductID.String(),
		ProductCode: p.ProductCode,
		Name:        p.Name,
		Quantity:    p.Quantity,
		UnitPrice:   p.UnitPrice,
	}
}

func toOrderServiceResponse(sv OrderService) OrderServiceResponse {
	return OrderServiceResponse{
		ID:           sv.ID.String(),
		ServiceID:    sv.ServiceID.String(),
		ServiceCode:  sv.ServiceCode,
		Name:         sv.Name,
		PriceApplied: sv.PriceApplied,
		Notes:        sv.Notes,
	}
}
