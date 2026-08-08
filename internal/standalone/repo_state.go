package standalone

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/state"
)

type sqliteStateRepo struct{ db *sql.DB }

func NewStateRepository(db *sql.DB) state.Repository {
	return &sqliteStateRepo{db: db}
}

func (r *sqliteStateRepo) List(ctx context.Context, name string, countryID *uuid.UUID, limit, offset int) ([]state.State, int64, error) {
	cid := ""
	if countryID != nil {
		cid = countryID.String()
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, code, country_id, created_at, updated_at, COUNT(*) OVER () AS total
		FROM states
		WHERE (? = '' OR name LIKE '%' || ? || '%')
		  AND (? = '' OR country_id = ?)
		ORDER BY name LIMIT ? OFFSET ?
	`, name, name, cid, cid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []state.State
	var total int64
	for rows.Next() {
		var s state.State
		var id, cID, cat, uat string
		if err := rows.Scan(&id, &s.Name, &s.Code, &cID, &cat, &uat, &total); err != nil {
			return nil, 0, err
		}
		s.ID, _ = uuid.Parse(id)
		s.CountryID, _ = uuid.Parse(cID)
		s.CreatedAt = parseTime(cat)
		s.UpdatedAt = parseTime(uat)
		result = append(result, s)
	}
	return result, total, rows.Err()
}

func (r *sqliteStateRepo) GetByID(ctx context.Context, id uuid.UUID) (*state.State, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, code, country_id, created_at, updated_at FROM states WHERE id = ?`, id.String())
	var s state.State
	var sid, cID, cat, uat string
	if err := row.Scan(&sid, &s.Name, &s.Code, &cID, &cat, &uat); err != nil {
		if err == sql.ErrNoRows {
			return nil, state.ErrNotFound
		}
		return nil, err
	}
	s.ID, _ = uuid.Parse(sid)
	s.CountryID, _ = uuid.Parse(cID)
	s.CreatedAt = parseTime(cat)
	s.UpdatedAt = parseTime(uat)
	return &s, nil
}

func (r *sqliteStateRepo) Create(ctx context.Context, s state.State) (*state.State, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `INSERT INTO states (id, name, code, country_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID.String(), s.Name, s.Code, s.CountryID.String(), now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, s.ID)
}

func (r *sqliteStateRepo) Update(ctx context.Context, s state.State) (*state.State, error) {
	_, err := r.db.ExecContext(ctx, `UPDATE states SET name=?, code=?, updated_at=? WHERE id=?`,
		s.Name, s.Code, nowStr(), s.ID.String())
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, s.ID)
}
