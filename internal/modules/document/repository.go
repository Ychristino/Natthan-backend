package document

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("documento não encontrado")

type Repository interface {
	Create(ctx context.Context, d Document) (*Document, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Document, error)
	List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]Document, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const cols = `id, person_id, type::text, value, created_at, updated_at`

func (r *repository) Create(ctx context.Context, d Document) (*Document, error) {
	var id pgtype.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO documents (id, person_id, type, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (type, value) DO UPDATE SET person_id = EXCLUDED.person_id
		RETURNING id
	`,
		pgtype.UUID{Bytes: d.ID, Valid: true},
		pgtype.UUID{Bytes: d.PersonID, Valid: true},
		d.Type, d.Value,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, uuid.UUID(id.Bytes))
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+cols+` FROM documents WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return scan(row)
}

func (r *repository) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]Document, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+cols+` FROM documents WHERE person_id = $1 ORDER BY type`,
		pgtype.UUID{Bytes: personID, Valid: true})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Document
	for rows.Next() {
		d, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, rows.Err()
}

func (r *repository) List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]Document, int64, error) {
	var pid *pgtype.UUID
	if personID != nil {
		v := pgtype.UUID{Bytes: *personID, Valid: true}
		pid = &v
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+cols+`
		FROM documents
		WHERE ($1::uuid IS NULL OR person_id = $1)
		ORDER BY type
		LIMIT $2 OFFSET $3
	`, pid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var total int64
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM documents WHERE ($1::uuid IS NULL OR person_id = $1)`, pid).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	var result []Document
	for rows.Next() {
		d, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *d)
	}
	return result, total, rows.Err()
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM documents WHERE id = $1`,
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

func scan(s scanner) (*Document, error) {
	var d Document
	var id, personID pgtype.UUID
	err := s.Scan(&id, &personID, &d.Type, &d.Value, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d.ID = uuid.UUID(id.Bytes)
	d.PersonID = uuid.UUID(personID.Bytes)
	return &d, nil
}
