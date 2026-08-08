package product

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID
	ProductCode   string
	Name          string
	Description   string
	PurchasePrice float64
	SalePrice     float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
