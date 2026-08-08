package standalone

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/person"
)

type sqlitePersonRepo struct{ db *sql.DB }

func NewPersonRepository(db *sql.DB) person.Repository {
	return &sqlitePersonRepo{db: db}
}

const personSelect = `
	SELECT p.id, p.full_name, p.birth_date, p.nationality_id,
	       c.name, c.code,
	       p.created_at, p.updated_at
	FROM persons p
	LEFT JOIN countries c ON c.id = p.nationality_id`

func (r *sqlitePersonRepo) Create(ctx context.Context, p person.Person) (*person.Person, error) {
	var nationalityID sql.NullString
	if p.NationalityID != nil {
		nationalityID = sql.NullString{String: p.NationalityID.String(), Valid: true}
	}
	var birthDate sql.NullString
	if p.BirthDate != nil {
		birthDate = sql.NullString{String: p.BirthDate.UTC().Format(time.RFC3339), Valid: true}
	}
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO persons (id, full_name, birth_date, nationality_id, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, ?, ?)`,
		p.ID.String(), p.FullName, birthDate, nationalityID, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, p.ID)
}

func (r *sqlitePersonRepo) List(ctx context.Context, name string, limit, offset int) ([]person.Person, int64, error) {
	rows, err := r.db.QueryContext(ctx, personSelect+`
		WHERE p.active = 1
		  AND (? = '' OR LOWER(p.full_name) LIKE LOWER('%' || ? || '%'))
		ORDER BY p.full_name
		LIMIT ? OFFSET ?
	`, name, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var persons []person.Person
	for rows.Next() {
		p, err := scanPersonRow(rows)
		if err != nil {
			return nil, 0, err
		}
		persons = append(persons, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM persons
		WHERE active = 1 AND (? = '' OR LOWER(full_name) LIKE LOWER('%' || ? || '%'))
	`, name, name).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return persons, total, nil
}

func (r *sqlitePersonRepo) GetByID(ctx context.Context, id uuid.UUID) (*person.Person, error) {
	row := r.db.QueryRowContext(ctx, personSelect+` WHERE p.id = ?`, id.String())
	return scanPersonRow(row)
}

func (r *sqlitePersonRepo) Update(ctx context.Context, p person.Person) (*person.Person, error) {
	var nationalityID sql.NullString
	if p.NationalityID != nil {
		nationalityID = sql.NullString{String: p.NationalityID.String(), Valid: true}
	}
	var birthDate sql.NullString
	if p.BirthDate != nil {
		birthDate = sql.NullString{String: p.BirthDate.UTC().Format(time.RFC3339), Valid: true}
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE persons SET full_name = ?, birth_date = ?, nationality_id = ?, updated_at = ?
		WHERE id = ?`,
		p.FullName, birthDate, nationalityID, nowStr(), p.ID.String())
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, person.ErrNotFound
	}
	return r.GetByID(ctx, p.ID)
}

func (r *sqlitePersonRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE persons SET active = 0, updated_at = ? WHERE id = ? AND active = 1`,
		nowStr(), id.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return person.ErrNotFound
	}
	return nil
}

func (r *sqlitePersonRepo) GetPersonByDocumentValue(ctx context.Context, value string) (*person.Person, error) {
	row := r.db.QueryRowContext(ctx, personSelect+`
		JOIN documents d ON d.person_id = p.id
		WHERE d.value = ?
		LIMIT 1
	`, value)
	p, err := scanPersonRow(row)
	if err != nil {
		if errors.Is(err, person.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPersonRow(s rowScanner) (*person.Person, error) {
	var p person.Person
	var id string
	var birthDate, nationalityID, nationalityName, nationalityCode sql.NullString
	var cat, uat string
	err := s.Scan(
		&id, &p.FullName, &birthDate, &nationalityID,
		&nationalityName, &nationalityCode,
		&cat, &uat,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, person.ErrNotFound
		}
		return nil, err
	}
	p.ID, _ = uuid.Parse(id)
	p.CreatedAt = parseTime(cat)
	p.UpdatedAt = parseTime(uat)
	if birthDate.Valid && birthDate.String != "" {
		t := parseTime(birthDate.String)
		p.BirthDate = &t
	}
	if nationalityID.Valid && nationalityID.String != "" {
		nid, _ := uuid.Parse(nationalityID.String)
		p.NationalityID = &nid
		p.NationalityName = nationalityName.String
		p.NationalityCode = nationalityCode.String
	}
	return &p, nil
}
