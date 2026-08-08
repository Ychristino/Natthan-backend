package address

import (
	"time"

	"github.com/google/uuid"
)

type Address struct {
	ID           uuid.UUID
	PersonID     uuid.UUID
	AddressType  string
	Street       string
	Number       string
	Complement   string
	Neighborhood string
	ZipCode      string
	CityID       uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AddressDetail struct {
	ID           uuid.UUID
	PersonID     uuid.UUID
	AddressType  string
	Street       string
	Number       string
	Complement   string
	Neighborhood string
	ZipCode      string
	CityID       uuid.UUID
	CityName     string
	StateID      uuid.UUID
	StateName    string
	StateCode    string
	CountryID    uuid.UUID
	CountryName  string
	CountryCode  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
