package contact

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID        uuid.UUID
	PersonID  uuid.UUID
	Type      string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
