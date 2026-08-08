package person

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("pessoa não encontrada")

type Repository interface {
	Create(ctx context.Context, p Person) (*Person, error)
	List(ctx context.Context, name string, limit, offset int) ([]Person, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Person, error)
	Update(ctx context.Context, p Person) (*Person, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetPersonByDocumentValue(ctx context.Context, value string) (*Person, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const personSelect = `
	SELECT p.id, p.full_name, p.birth_date, p.nationality_id,
	       c.name, c.code,
	       p.created_at, p.updated_at
	FROM persons p
	LEFT JOIN countries c ON c.id = p.nationality_id`

// ─── Person ───────────────────────────────────────────────────────────────────

func (r *repository) Create(ctx context.Context, p Person) (*Person, error) {
	var birthDate pgtype.Date
	if p.BirthDate != nil {
		birthDate = pgtype.Date{Time: *p.BirthDate, Valid: true}
	}
	var nationalityID pgtype.UUID
	if p.NationalityID != nil {
		nationalityID = pgtype.UUID{Bytes: *p.NationalityID, Valid: true}
	}

	var id pgtype.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO persons (id, full_name, birth_date, nationality_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, pgtype.UUID{Bytes: p.ID, Valid: true}, p.FullName, birthDate, nationalityID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, uuid.UUID(id.Bytes))
}

func (r *repository) List(ctx context.Context, name string, limit, offset int) ([]Person, int64, error) {
	rows, err := r.db.Query(ctx, personSelect+`
		WHERE p.active = true
		  AND ($1 = '' OR p.full_name ILIKE '%' || $1 || '%')
		ORDER BY p.full_name
		LIMIT $2 OFFSET $3
	`, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, 0, err
		}
		persons = append(persons, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM persons WHERE active = true AND ($1 = '' OR full_name ILIKE '%' || $1 || '%')
	`, name).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return persons, total, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	row := r.db.QueryRow(ctx, personSelect+` WHERE p.id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return scanPerson(row)
}

func (r *repository) Update(ctx context.Context, p Person) (*Person, error) {
	var birthDate pgtype.Date
	if p.BirthDate != nil {
		birthDate = pgtype.Date{Time: *p.BirthDate, Valid: true}
	}
	var nationalityID pgtype.UUID
	if p.NationalityID != nil {
		nationalityID = pgtype.UUID{Bytes: *p.NationalityID, Valid: true}
	}

	tag, err := r.db.Exec(ctx, `
		UPDATE persons SET full_name = $2, birth_date = $3, nationality_id = $4, updated_at = NOW()
		WHERE id = $1
	`, pgtype.UUID{Bytes: p.ID, Valid: true}, p.FullName, birthDate, nationalityID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, p.ID)
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE persons SET active = false, updated_at = NOW() WHERE id = $1 AND active = true`,
		pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── Document lookup ──────────────────────────────────────────────────────────

func (r *repository) GetPersonByDocumentValue(ctx context.Context, value string) (*Person, error) {
	row := r.db.QueryRow(ctx, personSelect+`
		JOIN documents d ON d.person_id = p.id
		WHERE d.value = $1
		LIMIT 1
	`, value)
	p, err := scanPerson(row)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanPerson(s scanner) (*Person, error) {
	var p Person
	var id, nationalityID pgtype.UUID
	var birthDate pgtype.Date
	var nationalityName, nationalityCode pgtype.Text

	err := s.Scan(
		&id, &p.FullName, &birthDate, &nationalityID,
		&nationalityName, &nationalityCode,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	p.ID = uuid.UUID(id.Bytes)
	if birthDate.Valid {
		p.BirthDate = &birthDate.Time
	}
	if nationalityID.Valid {
		v := uuid.UUID(nationalityID.Bytes)
		p.NationalityID = &v
		p.NationalityName = nationalityName.String
		p.NationalityCode = nationalityCode.String
	}
	return &p, nil
}
