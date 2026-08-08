package state

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("estado não encontrado")

type Repository interface {
	List(ctx context.Context, name string, countryID *uuid.UUID, limit, offset int) ([]State, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*State, error)
	Create(ctx context.Context, s State) (*State, error)
	Update(ctx context.Context, s State) (*State, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const stateColumns = `id, name, code, country_id, created_at, updated_at`

func (r *repository) List(ctx context.Context, name string, countryID *uuid.UUID, limit, offset int) ([]State, int64, error) {
	var cid *pgtype.UUID
	if countryID != nil {
		v := pgtype.UUID{Bytes: *countryID, Valid: true}
		cid = &v
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+stateColumns+`, COUNT(*) OVER () AS total
		FROM states
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		  AND ($2::uuid IS NULL OR country_id = $2)
		ORDER BY name
		LIMIT $3 OFFSET $4
	`, name, cid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var states []State
	var total int64
	for rows.Next() {
		s, t, err := scanWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		states = append(states, *s)
	}
	return states, total, rows.Err()
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*State, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+stateColumns+` FROM states WHERE id = $1
	`, pgtype.UUID{Bytes: id, Valid: true})
	return scan(row)
}

func (r *repository) Create(ctx context.Context, s State) (*State, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO states (id, name, code, country_id) VALUES ($1, $2, $3, $4)
		RETURNING `+stateColumns,
		pgtype.UUID{Bytes: s.ID, Valid: true}, s.Name, s.Code,
		pgtype.UUID{Bytes: s.CountryID, Valid: true},
	)
	return scan(row)
}

func (r *repository) Update(ctx context.Context, s State) (*State, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE states SET name = $2, code = $3, updated_at = NOW()
		WHERE id = $1 RETURNING `+stateColumns,
		pgtype.UUID{Bytes: s.ID, Valid: true}, s.Name, s.Code,
	)
	return scan(row)
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (*State, error) {
	var st State
	var id, countryID pgtype.UUID
	err := s.Scan(&id, &st.Name, &st.Code, &countryID, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	st.ID = uuid.UUID(id.Bytes)
	st.CountryID = uuid.UUID(countryID.Bytes)
	return &st, nil
}

func scanWithTotal(s scanner) (*State, int64, error) {
	var st State
	var id, countryID pgtype.UUID
	var total int64
	err := s.Scan(&id, &st.Name, &st.Code, &countryID, &st.CreatedAt, &st.UpdatedAt, &total)
	if err != nil {
		return nil, 0, err
	}
	st.ID = uuid.UUID(id.Bytes)
	st.CountryID = uuid.UUID(countryID.Bytes)
	return &st, total, nil
}
