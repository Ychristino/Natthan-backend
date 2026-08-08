package standalone

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/country"
)

type sqliteCountryRepo struct{ db *sql.DB }

func NewCountryRepository(db *sql.DB) country.Repository {
	return &sqliteCountryRepo{db: db}
}

func (r *sqliteCountryRepo) List(ctx context.Context, name string, limit, offset int) ([]country.Country, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, code, created_at, updated_at, COUNT(*) OVER () AS total
		FROM countries
		WHERE (? = '' OR name LIKE '%' || ? || '%')
		ORDER BY name LIMIT ? OFFSET ?
	`, name, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []country.Country
	var total int64
	for rows.Next() {
		var c country.Country
		var id, cat, uat string
		if err := rows.Scan(&id, &c.Name, &c.Code, &cat, &uat, &total); err != nil {
			return nil, 0, err
		}
		c.ID, _ = uuid.Parse(id)
		c.CreatedAt = parseTime(cat)
		c.UpdatedAt = parseTime(uat)
		result = append(result, c)
	}
	return result, total, rows.Err()
}

func (r *sqliteCountryRepo) GetByID(ctx context.Context, id uuid.UUID) (*country.Country, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, code, created_at, updated_at FROM countries WHERE id = ?`, id.String())
	var c country.Country
	var sid, cat, uat string
	if err := row.Scan(&sid, &c.Name, &c.Code, &cat, &uat); err != nil {
		if err == sql.ErrNoRows {
			return nil, country.ErrNotFound
		}
		return nil, err
	}
	c.ID, _ = uuid.Parse(sid)
	c.CreatedAt = parseTime(cat)
	c.UpdatedAt = parseTime(uat)
	return &c, nil
}

func (r *sqliteCountryRepo) Create(ctx context.Context, c country.Country) (*country.Country, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `INSERT INTO countries (id, name, code, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		c.ID.String(), c.Name, c.Code, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, c.ID)
}

func (r *sqliteCountryRepo) Update(ctx context.Context, c country.Country) (*country.Country, error) {
	_, err := r.db.ExecContext(ctx, `UPDATE countries SET name=?, code=?, updated_at=? WHERE id=?`,
		c.Name, c.Code, nowStr(), c.ID.String())
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, c.ID)
}
