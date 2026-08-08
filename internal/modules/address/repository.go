package address

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("endereço não encontrado")

type Repository interface {
	Create(ctx context.Context, a Address) (*AddressDetail, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AddressDetail, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) ([]AddressDetail, error)
	List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]AddressDetail, int64, error)
	Update(ctx context.Context, a Address) (*AddressDetail, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const detailSelect = `
	SELECT
		a.id, a.person_id, a.address_type::text, a.street, a.number, a.complement,
		a.neighborhood, a.zip_code, a.created_at, a.updated_at,
		ci.id, ci.name,
		s.id, s.name, s.code,
		co.id, co.name, co.code
	FROM addresses a
	JOIN cities   ci ON ci.id = a.city_id
	JOIN states   s  ON s.id  = ci.state_id
	JOIN countries co ON co.id = s.country_id`

func (r *repository) Create(ctx context.Context, a Address) (*AddressDetail, error) {
	var complement pgtype.Text
	if a.Complement != "" {
		complement = pgtype.Text{String: a.Complement, Valid: true}
	}

	var id pgtype.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO addresses (id, person_id, address_type, street, number, complement, neighborhood, zip_code, city_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`,
		pgtype.UUID{Bytes: a.ID, Valid: true},
		pgtype.UUID{Bytes: a.PersonID, Valid: true},
		a.AddressType,
		a.Street, a.Number, complement, a.Neighborhood, a.ZipCode,
		pgtype.UUID{Bytes: a.CityID, Valid: true},
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, uuid.UUID(id.Bytes))
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*AddressDetail, error) {
	row := r.db.QueryRow(ctx, detailSelect+` WHERE a.id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return scanDetail(row)
}

func (r *repository) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]AddressDetail, error) {
	rows, err := r.db.Query(ctx, detailSelect+` WHERE a.person_id = $1 ORDER BY a.created_at`,
		pgtype.UUID{Bytes: personID, Valid: true})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AddressDetail
	for rows.Next() {
		d, err := scanDetail(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, rows.Err()
}

func (r *repository) List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]AddressDetail, int64, error) {
	var pid *pgtype.UUID
	if personID != nil {
		v := pgtype.UUID{Bytes: *personID, Valid: true}
		pid = &v
	}

	rows, err := r.db.Query(ctx, detailSelect+`
		WHERE ($1::uuid IS NULL OR a.person_id = $1)
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`, pid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// get total separately to avoid complexity with window functions + joins
	var total int64
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM addresses WHERE ($1::uuid IS NULL OR person_id = $1)
	`, pid).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	var result []AddressDetail
	for rows.Next() {
		d, err := scanDetail(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *d)
	}
	return result, total, rows.Err()
}

func (r *repository) Update(ctx context.Context, a Address) (*AddressDetail, error) {
	var complement pgtype.Text
	if a.Complement != "" {
		complement = pgtype.Text{String: a.Complement, Valid: true}
	}

	tag, err := r.db.Exec(ctx, `
		UPDATE addresses
		SET address_type = $2, street = $3, number = $4, complement = $5,
		    neighborhood = $6, zip_code = $7, city_id = $8, updated_at = NOW()
		WHERE id = $1
	`,
		pgtype.UUID{Bytes: a.ID, Valid: true},
		a.AddressType,
		a.Street, a.Number, complement, a.Neighborhood, a.ZipCode,
		pgtype.UUID{Bytes: a.CityID, Valid: true},
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, a.ID)
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM addresses WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanDetail(s scanner) (*AddressDetail, error) {
	var d AddressDetail
	var (
		id, personID, cityID, stateID, countryID pgtype.UUID
		comp                                     pgtype.Text
	)
	err := s.Scan(
		&id, &personID, &d.AddressType, &d.Street, &d.Number, &comp,
		&d.Neighborhood, &d.ZipCode, &d.CreatedAt, &d.UpdatedAt,
		&cityID, &d.CityName,
		&stateID, &d.StateName, &d.StateCode,
		&countryID, &d.CountryName, &d.CountryCode,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d.ID = uuid.UUID(id.Bytes)
	d.PersonID = uuid.UUID(personID.Bytes)
	d.CityID = uuid.UUID(cityID.Bytes)
	d.StateID = uuid.UUID(stateID.Bytes)
	d.CountryID = uuid.UUID(countryID.Bytes)
	d.Complement = comp.String
	return &d, nil
}
