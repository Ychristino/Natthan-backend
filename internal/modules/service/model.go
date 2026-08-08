package service

import (
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID          uuid.UUID
	ServiceCode string
	Name        string
	Description string
	Price       float64
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
