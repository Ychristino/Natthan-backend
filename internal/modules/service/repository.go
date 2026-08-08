package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("serviço não encontrado")

type Repository interface {
	Create(ctx context.Context, s Service) (*Service, error)
	List(ctx context.Context, name string, limit, offset int) ([]Service, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Service, error)
	Update(ctx context.Context, s Service) (*Service, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const serviceColumns = `id, service_code, name, description, price, active, created_at, updated_at`

func (r *repository) Create(ctx context.Context, s Service) (*Service, error) {
	var desc pgtype.Text
	if s.Description != "" {
		desc = pgtype.Text{String: s.Description, Valid: true}
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO services (id, service_code, name, description, price, active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING `+serviceColumns,
		pgtype.UUID{Bytes: s.ID, Valid: true},
		s.ServiceCode, s.Name, desc, s.Price,
	)
	return scanService(row)
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Service, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+serviceColumns+` FROM services WHERE id = $1 AND active = true
	`, pgtype.UUID{Bytes: id, Valid: true})
	return scanService(row)
}

func (r *repository) List(ctx context.Context, name string, limit, offset int) ([]Service, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+serviceColumns+`, COUNT(*) OVER () AS total
		FROM services
		WHERE active = true
		  AND ($1 = '' OR name ILIKE '%' || $1 || '%')
		ORDER BY name
		LIMIT $2 OFFSET $3
	`, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var services []Service
	var total int64
	for rows.Next() {
		s, t, err := scanServiceWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		services = append(services, *s)
	}
	return services, total, rows.Err()
}

func (r *repository) Update(ctx context.Context, s Service) (*Service, error) {
	var desc pgtype.Text
	if s.Description != "" {
		desc = pgtype.Text{String: s.Description, Valid: true}
	}
	row := r.db.QueryRow(ctx, `
		UPDATE services
		SET service_code = $2, name = $3, description = $4, price = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING `+serviceColumns,
		pgtype.UUID{Bytes: s.ID, Valid: true},
		s.ServiceCode, s.Name, desc, s.Price,
	)
	return scanService(row)
}

func (r *repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE services SET active = false, updated_at = NOW() WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return err
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanService(s scanner) (*Service, error) {
	var svc Service
	var id pgtype.UUID
	var desc pgtype.Text
	err := s.Scan(&id, &svc.ServiceCode, &svc.Name, &desc, &svc.Price, &svc.Active, &svc.CreatedAt, &svc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	svc.ID = uuid.UUID(id.Bytes)
	svc.Description = desc.String
	return &svc, nil
}

func scanServiceWithTotal(s scanner) (*Service, int64, error) {
	var svc Service
	var id pgtype.UUID
	var desc pgtype.Text
	var total int64
	err := s.Scan(&id, &svc.ServiceCode, &svc.Name, &desc, &svc.Price, &svc.Active, &svc.CreatedAt, &svc.UpdatedAt, &total)
	if err != nil {
		return nil, 0, err
	}
	svc.ID = uuid.UUID(id.Bytes)
	svc.Description = desc.String
	return &svc, total, nil
}
