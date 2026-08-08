package country

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("país não encontrado")

type Repository interface {
	List(ctx context.Context, name string, limit, offset int) ([]Country, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Country, error)
	Create(ctx context.Context, c Country) (*Country, error)
	Update(ctx context.Context, c Country) (*Country, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const countryColumns = `id, name, code, created_at, updated_at`

func (r *repository) List(ctx context.Context, name string, limit, offset int) ([]Country, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+countryColumns+`, COUNT(*) OVER () AS total
		FROM countries
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		ORDER BY name
		LIMIT $2 OFFSET $3
	`, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var countries []Country
	var total int64
	for rows.Next() {
		c, t, err := scanWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		countries = append(countries, *c)
	}
	return countries, total, rows.Err()
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Country, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+countryColumns+` FROM countries WHERE id = $1
	`, pgtype.UUID{Bytes: id, Valid: true})
	return scan(row)
}

func (r *repository) Create(ctx context.Context, c Country) (*Country, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO countries (id, name, code) VALUES ($1, $2, $3)
		RETURNING `+countryColumns,
		pgtype.UUID{Bytes: c.ID, Valid: true}, c.Name, c.Code,
	)
	return scan(row)
}

func (r *repository) Update(ctx context.Context, c Country) (*Country, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE countries SET name = $2, code = $3, updated_at = NOW()
		WHERE id = $1 RETURNING `+countryColumns,
		pgtype.UUID{Bytes: c.ID, Valid: true}, c.Name, c.Code,
	)
	return scan(row)
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (*Country, error) {
	var c Country
	var id pgtype.UUID
	err := s.Scan(&id, &c.Name, &c.Code, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c.ID = uuid.UUID(id.Bytes)
	return &c, nil
}

func scanWithTotal(s scanner) (*Country, int64, error) {
	var c Country
	var id pgtype.UUID
	var total int64
	err := s.Scan(&id, &c.Name, &c.Code, &c.CreatedAt, &c.UpdatedAt, &total)
	if err != nil {
		return nil, 0, err
	}
	c.ID = uuid.UUID(id.Bytes)
	return &c, total, nil
}
