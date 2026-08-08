package state

import (
	"time"

	"github.com/google/uuid"
)

type State struct {
	ID        uuid.UUID
	Name      string
	Code      string
	CountryID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
