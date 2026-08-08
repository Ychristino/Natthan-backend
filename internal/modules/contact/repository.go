package contact

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("contato não encontrado")

type Repository interface {
	Create(ctx context.Context, c Contact) (*Contact, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Contact, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Contact, error)
	List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]Contact, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, c Contact) (*Contact, error) {
	var id pgtype.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO contacts (id, person_id, type, value)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		pgtype.UUID{Bytes: c.ID, Valid: true},
		pgtype.UUID{Bytes: c.PersonID, Valid: true},
		c.Type, c.Value,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, uuid.UUID(id.Bytes))
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Contact, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, person_id, type::text, value, created_at, updated_at FROM contacts WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return scan(row)
}

func (r *repository) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Contact, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, person_id, type::text, value, created_at, updated_at FROM contacts WHERE person_id = $1 ORDER BY created_at`,
		pgtype.UUID{Bytes: personID, Valid: true})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Contact
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *c)
	}
	return result, rows.Err()
}

func (r *repository) List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]Contact, int64, error) {
	var pid *pgtype.UUID
	if personID != nil {
		v := pgtype.UUID{Bytes: *personID, Valid: true}
		pid = &v
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, person_id, type::text, value, created_at, updated_at
		FROM contacts
		WHERE ($1::uuid IS NULL OR person_id = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, pid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var total int64
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM contacts WHERE ($1::uuid IS NULL OR person_id = $1)`, pid).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	var result []Contact
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *c)
	}
	return result, total, rows.Err()
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM contacts WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (*Contact, error) {
	var c Contact
	var id, personID pgtype.UUID
	err := s.Scan(&id, &personID, &c.Type, &c.Value, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c.ID = uuid.UUID(id.Bytes)
	c.PersonID = uuid.UUID(personID.Bytes)
	return &c, nil
}
