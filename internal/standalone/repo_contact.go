package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/contact"
)

type sqliteContactRepo struct{ db *sql.DB }

func NewContactRepository(db *sql.DB) contact.Repository {
	return &sqliteContactRepo{db: db}
}

func (r *sqliteContactRepo) Create(ctx context.Context, c contact.Contact) (*contact.Contact, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO contacts (id, person_id, type, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID.String(), c.PersonID.String(), c.Type, c.Value, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, c.ID)
}

func (r *sqliteContactRepo) GetByID(ctx context.Context, id uuid.UUID) (*contact.Contact, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, person_id, type, value, created_at, updated_at FROM contacts WHERE id = ?`, id.String())
	return scanContact(row)
}

func (r *sqliteContactRepo) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]contact.Contact, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, person_id, type, value, created_at, updated_at FROM contacts WHERE person_id = ? ORDER BY created_at`,
		personID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []contact.Contact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *c)
	}
	return result, rows.Err()
}

func (r *sqliteContactRepo) List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]contact.Contact, int64, error) {
	pid := ""
	if personID != nil {
		pid = personID.String()
	}

	var total int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM contacts WHERE (? = '' OR person_id = ?)`, pid, pid).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, person_id, type, value, created_at, updated_at
		FROM contacts
		WHERE (? = '' OR person_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, pid, pid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []contact.Contact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *c)
	}
	return result, total, rows.Err()
}

func (r *sqliteContactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM contacts WHERE id = ?`, id.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return contact.ErrNotFound
	}
	return nil
}

func scanContact(s rowScanner) (*contact.Contact, error) {
	var c contact.Contact
	var id, personID, cat, uat string
	err := s.Scan(&id, &personID, &c.Type, &c.Value, &cat, &uat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, contact.ErrNotFound
		}
		return nil, err
	}
	c.ID, _ = uuid.Parse(id)
	c.PersonID, _ = uuid.Parse(personID)
	c.CreatedAt = parseTime(cat)
	c.UpdatedAt = parseTime(uat)
	return &c, nil
}
