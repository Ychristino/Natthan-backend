package person

import (
	"time"

	"github.com/google/uuid"
)

type Person struct {
	ID              uuid.UUID
	FullName        string
	BirthDate       *time.Time
	NationalityID   *uuid.UUID
	NationalityName string
	NationalityCode string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type LookupResult struct {
	Found    bool
	PersonID uuid.UUID
	FullName string
}
