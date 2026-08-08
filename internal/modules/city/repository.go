package city

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("cidade não encontrada")

type Repository interface {
	List(ctx context.Context, name string, stateID *uuid.UUID, limit, offset int) ([]City, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*City, error)
	Create(ctx context.Context, c City) (*City, error)
	Update(ctx context.Context, c City) (*City, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const cityColumns = `id, name, state_id, created_at, updated_at`

func (r *repository) List(ctx context.Context, name string, stateID *uuid.UUID, limit, offset int) ([]City, int64, error) {
	var sid *pgtype.UUID
	if stateID != nil {
		v := pgtype.UUID{Bytes: *stateID, Valid: true}
		sid = &v
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+cityColumns+`, COUNT(*) OVER () AS total
		FROM cities
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		  AND ($2::uuid IS NULL OR state_id = $2)
		ORDER BY name
		LIMIT $3 OFFSET $4
	`, name, sid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var cities []City
	var total int64
	for rows.Next() {
		c, t, err := scanWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		cities = append(cities, *c)
	}
	return cities, total, rows.Err()
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*City, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+cityColumns+` FROM cities WHERE id = $1
	`, pgtype.UUID{Bytes: id, Valid: true})
	return scan(row)
}

func (r *repository) Create(ctx context.Context, c City) (*City, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO cities (id, name, state_id) VALUES ($1, $2, $3)
		RETURNING `+cityColumns,
		pgtype.UUID{Bytes: c.ID, Valid: true}, c.Name,
		pgtype.UUID{Bytes: c.StateID, Valid: true},
	)
	return scan(row)
}

func (r *repository) Update(ctx context.Context, c City) (*City, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE cities SET name = $2, updated_at = NOW()
		WHERE id = $1 RETURNING `+cityColumns,
		pgtype.UUID{Bytes: c.ID, Valid: true}, c.Name,
	)
	return scan(row)
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (*City, error) {
	var c City
	var id, stateID pgtype.UUID
	err := s.Scan(&id, &c.Name, &stateID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c.ID = uuid.UUID(id.Bytes)
	c.StateID = uuid.UUID(stateID.Bytes)
	return &c, nil
}

func scanWithTotal(s scanner) (*City, int64, error) {
	var c City
	var id, stateID pgtype.UUID
	var total int64
	err := s.Scan(&id, &c.Name, &stateID, &c.CreatedAt, &c.UpdatedAt, &total)
	if err != nil {
		return nil, 0, err
	}
	c.ID = uuid.UUID(id.Bytes)
	c.StateID = uuid.UUID(stateID.Bytes)
	return &c, total, nil
}
