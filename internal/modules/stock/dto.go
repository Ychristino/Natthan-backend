package stock

import "time"

// ─── Requests ─────────────────────────────────────────────────────────────────

type ListStockRequest struct {
	Name  string `query:"name"`
	Page  int    `query:"page"  validate:"min=1"`
	Limit int    `query:"limit" validate:"min=1,max=100"`
}

type ListMovementsRequest struct {
	ProductID string `query:"product_id"`
	Type      string `query:"type"       validate:"omitempty,oneof=in out adjustment"`
	Page      int    `query:"page"       validate:"min=1"`
	Limit     int    `query:"limit"      validate:"min=1,max=100"`
}

type CreateMovementRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Type      string `json:"type"       validate:"required,oneof=in out adjustment"`
	Quantity  int    `json:"quantity"   validate:"required,min=1"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type StockResponse struct {
	ID             string  `json:"id"`
	ProductID      string  `json:"product_id"`
	ProductCode    string  `json:"product_code"`
	ProductName    string  `json:"product_name"`
	Quantity       int     `json:"quantity"`
	LastMovementAt *string `json:"last_movement_at,omitempty"`
	UpdatedAt      string  `json:"updated_at"`
}

type MovementResponse struct {
	ID                string  `json:"id"`
	ProductID         string  `json:"product_id"`
	ProductCode       string  `json:"product_code"`
	ProductName       string  `json:"product_name"`
	Type              string  `json:"type"`
	Quantity          int     `json:"quantity"`
	ServiceOrderID    *string `json:"service_order_id,omitempty"`
	CreatedByUsername string  `json:"created_by_username"`
	CreatedAt         string  `json:"created_at"`
}

type StockListResponse struct {
	Data  []StockResponse `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

type MovementListResponse struct {
	Data  []MovementResponse `json:"data"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toStockResponse(s StockItem) StockResponse {
	r := StockResponse{
		ID:          s.ID.String(),
		ProductID:   s.ProductID.String(),
		ProductCode: s.ProductCode,
		ProductName: s.ProductName,
		Quantity:    s.Quantity,
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
	if s.LastMovementAt != nil {
		t := s.LastMovementAt.Format(time.RFC3339)
		r.LastMovementAt = &t
	}
	return r
}

func toMovementResponse(m Movement) MovementResponse {
	r := MovementResponse{
		ID:                m.ID.String(),
		ProductID:         m.ProductID.String(),
		ProductCode:       m.ProductCode,
		ProductName:       m.ProductName,
		Type:              string(m.Type),
		Quantity:          m.Quantity,
		CreatedByUsername: m.CreatedByUsername,
		CreatedAt:         m.CreatedAt.Format(time.RFC3339),
	}
	if m.ServiceOrderID != nil {
		s := m.ServiceOrderID.String()
		r.ServiceOrderID = &s
	}
	return r
}
