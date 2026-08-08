package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/document"
)

type sqliteDocumentRepo struct{ db *sql.DB }

func NewDocumentRepository(db *sql.DB) document.Repository {
	return &sqliteDocumentRepo{db: db}
}

func (r *sqliteDocumentRepo) Create(ctx context.Context, d document.Document) (*document.Document, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO documents (id, person_id, type, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (type, value) DO UPDATE SET person_id = excluded.person_id`,
		d.ID.String(), d.PersonID.String(), d.Type, d.Value, now, now)
	if err != nil {
		return nil, err
	}
	// Fetch the actual row (may have been updated by ON CONFLICT).
	row := r.db.QueryRowContext(ctx,
		`SELECT id, person_id, type, value, created_at, updated_at FROM documents WHERE type = ? AND value = ?`,
		d.Type, d.Value)
	return scanDocument(row)
}

func (r *sqliteDocumentRepo) GetByID(ctx context.Context, id uuid.UUID) (*document.Document, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, person_id, type, value, created_at, updated_at FROM documents WHERE id = ?`, id.String())
	return scanDocument(row)
}

func (r *sqliteDocumentRepo) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]document.Document, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, person_id, type, value, created_at, updated_at FROM documents WHERE person_id = ? ORDER BY type`,
		personID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []document.Document
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, rows.Err()
}

func (r *sqliteDocumentRepo) List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]document.Document, int64, error) {
	pid := ""
	if personID != nil {
		pid = personID.String()
	}

	var total int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents WHERE (? = '' OR person_id = ?)`, pid, pid).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, person_id, type, value, created_at, updated_at
		FROM documents
		WHERE (? = '' OR person_id = ?)
		ORDER BY type
		LIMIT ? OFFSET ?
	`, pid, pid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []document.Document
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *d)
	}
	return result, total, rows.Err()
}

func (r *sqliteDocumentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return document.ErrNotFound
	}
	return nil
}

func scanDocument(s rowScanner) (*document.Document, error) {
	var d document.Document
	var id, personID, cat, uat string
	err := s.Scan(&id, &personID, &d.Type, &d.Value, &cat, &uat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, document.ErrNotFound
		}
		return nil, err
	}
	d.ID, _ = uuid.Parse(id)
	d.PersonID, _ = uuid.Parse(personID)
	d.CreatedAt = parseTime(cat)
	d.UpdatedAt = parseTime(uat)
	return &d, nil
}
