package standalone

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/city"
)

type sqliteCityRepo struct{ db *sql.DB }

func NewCityRepository(db *sql.DB) city.Repository {
	return &sqliteCityRepo{db: db}
}

func (r *sqliteCityRepo) List(ctx context.Context, name string, stateID *uuid.UUID, limit, offset int) ([]city.City, int64, error) {
	sid := ""
	if stateID != nil {
		sid = stateID.String()
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, state_id, created_at, updated_at, COUNT(*) OVER () AS total
		FROM cities
		WHERE (? = '' OR name LIKE '%' || ? || '%')
		  AND (? = '' OR state_id = ?)
		ORDER BY name LIMIT ? OFFSET ?
	`, name, name, sid, sid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []city.City
	var total int64
	for rows.Next() {
		var c city.City
		var id, sID, cat, uat string
		if err := rows.Scan(&id, &c.Name, &sID, &cat, &uat, &total); err != nil {
			return nil, 0, err
		}
		c.ID, _ = uuid.Parse(id)
		c.StateID, _ = uuid.Parse(sID)
		c.CreatedAt = parseTime(cat)
		c.UpdatedAt = parseTime(uat)
		result = append(result, c)
	}
	return result, total, rows.Err()
}

func (r *sqliteCityRepo) GetByID(ctx context.Context, id uuid.UUID) (*city.City, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, state_id, created_at, updated_at FROM cities WHERE id = ?`, id.String())
	var c city.City
	var sid, stID, cat, uat string
	if err := row.Scan(&sid, &c.Name, &stID, &cat, &uat); err != nil {
		if err == sql.ErrNoRows {
			return nil, city.ErrNotFound
		}
		return nil, err
	}
	c.ID, _ = uuid.Parse(sid)
	c.StateID, _ = uuid.Parse(stID)
	c.CreatedAt = parseTime(cat)
	c.UpdatedAt = parseTime(uat)
	return &c, nil
}

func (r *sqliteCityRepo) Create(ctx context.Context, c city.City) (*city.City, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `INSERT INTO cities (id, name, state_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		c.ID.String(), c.Name, c.StateID.String(), now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, c.ID)
}

func (r *sqliteCityRepo) Update(ctx context.Context, c city.City) (*city.City, error) {
	_, err := r.db.ExecContext(ctx, `UPDATE cities SET name=?, updated_at=? WHERE id=?`, c.Name, nowStr(), c.ID.String())
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, c.ID)
}
