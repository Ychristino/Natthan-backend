package product

import "time"

// ─── Requests ─────────────────────────────────────────────────────────────────

type CreateRequest struct {
	ProductCode   string  `json:"product_code"   validate:"required,max=100"`
	Name          string  `json:"name"           validate:"required,min=2,max=255"`
	Description   string  `json:"description"    validate:"max=1000"`
	PurchasePrice float64 `json:"purchase_price" validate:"gte=0"`
	SalePrice     float64 `json:"sale_price"     validate:"required,gt=0"`
}

type UpdateRequest struct {
	ProductCode   *string  `json:"product_code"   validate:"omitempty,max=100"`
	Name          *string  `json:"name"           validate:"omitempty,min=2,max=255"`
	Description   *string  `json:"description"    validate:"omitempty,max=1000"`
	PurchasePrice *float64 `json:"purchase_price" validate:"omitempty,gte=0"`
	SalePrice     *float64 `json:"sale_price"     validate:"omitempty,gt=0"`
}

type ListRequest struct {
	Name  string `query:"name"`
	Page  int    `query:"page"  validate:"min=1"`
	Limit int    `query:"limit" validate:"min=1,max=100"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

type Response struct {
	ID            string  `json:"id"`
	ProductCode   string  `json:"product_code"`
	Name          string  `json:"name"`
	Description   string  `json:"description,omitempty"`
	PurchasePrice float64 `json:"purchase_price"`
	SalePrice     float64 `json:"sale_price"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type ListResponse struct {
	Data  []Response `json:"data"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toResponse(p Product) Response {
	return Response{
		ID:            p.ID.String(),
		ProductCode:   p.ProductCode,
		Name:          p.Name,
		Description:   p.Description,
		PurchasePrice: p.PurchasePrice,
		SalePrice:     p.SalePrice,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}
