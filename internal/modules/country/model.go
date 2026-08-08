package country

import (
	"time"

	"github.com/google/uuid"
)

type Country struct {
	ID        uuid.UUID
	Name      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
