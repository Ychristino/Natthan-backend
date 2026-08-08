package city

import (
	"time"

	"github.com/google/uuid"
)

type City struct {
	ID        uuid.UUID
	Name      string
	StateID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
