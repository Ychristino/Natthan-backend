package stock

import (
	"time"

	"github.com/google/uuid"
)

type MovementType string

const (
	MovementIn         MovementType = "in"
	MovementOut        MovementType = "out"
	MovementAdjustment MovementType = "adjustment"
)

type StockItem struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	ProductCode    string
	ProductName    string
	Quantity       int
	LastMovementAt *time.Time
	UpdatedAt      time.Time
}

type Movement struct {
	ID                uuid.UUID
	ProductID         uuid.UUID
	ProductCode       string
	ProductName       string
	Type              MovementType
	Quantity          int
	ServiceOrderID    *uuid.UUID
	CreatedBy         uuid.UUID
	CreatedByUsername string
	CreatedAt         time.Time
}
