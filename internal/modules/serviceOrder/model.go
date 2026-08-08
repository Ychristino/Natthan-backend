package serviceOrder

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type OrderProduct struct {
	ID             uuid.UUID
	ServiceOrderID uuid.UUID
	ProductID      uuid.UUID
	ProductCode    string
	Name           string
	Quantity       int
	UnitPrice      float64
	CreatedAt      time.Time
}

type OrderService struct {
	ID             uuid.UUID
	ServiceOrderID uuid.UUID
	ServiceID      uuid.UUID
	ServiceCode    string
	Name           string
	PriceApplied   float64
	Notes          string
	CreatedAt      time.Time
}

type ServiceOrder struct {
	ID               uuid.UUID
	ClientID         uuid.UUID
	ClientName       string
	EmployeeID       uuid.UUID
	EmployeeFullName string
	Status           Status
	EntryDate        time.Time
	ExitDate         *time.Time
	PayDate          *time.Time
	IsPaid           bool
	Notes            string
	Products         []OrderProduct
	Services         []OrderService
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
