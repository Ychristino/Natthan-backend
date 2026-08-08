package document

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID        uuid.UUID
	PersonID  uuid.UUID
	Type      string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
