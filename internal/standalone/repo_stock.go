package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/stock"
)

type sqliteStockRepo struct{ db *sql.DB }

func NewStockRepository(db *sql.DB) stock.Repository {
	return &sqliteStockRepo{db: db}
}

func (r *sqliteStockRepo) List(ctx context.Context, name string, limit, offset int) ([]stock.StockItem, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.product_id, pr.product_code, pr.name, s.quantity,
		       MAX(sm.created_at) AS last_movement_at,
		       s.updated_at,
		       COUNT(*) OVER () AS total
		FROM stock s
		JOIN products pr ON pr.id = s.product_id
		LEFT JOIN stock_movements sm ON sm.product_id = s.product_id
		WHERE (? = '' OR LOWER(pr.name) LIKE LOWER('%' || ? || '%') OR LOWER(pr.product_code) LIKE LOWER('%' || ? || '%'))
		GROUP BY s.id, s.product_id, pr.product_code, pr.name, s.quantity, s.updated_at
		ORDER BY pr.name
		LIMIT ? OFFSET ?
	`, name, name, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []stock.StockItem
	var total int64
	for rows.Next() {
		var item stock.StockItem
		var id, productID string
		var lastMovAt sql.NullString
		var uat string
		var t int64
		if err := rows.Scan(&id, &productID, &item.ProductCode, &item.ProductName, &item.Quantity,
			&lastMovAt, &uat, &t); err != nil {
			return nil, 0, err
		}
		item.ID, _ = uuid.Parse(id)
		item.ProductID, _ = uuid.Parse(productID)
		item.UpdatedAt = parseTime(uat)
		if lastMovAt.Valid && lastMovAt.String != "" {
			t2 := parseTime(lastMovAt.String)
			item.LastMovementAt = &t2
		}
		total = t
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *sqliteStockRepo) ListMovements(ctx context.Context, productID, movType string, limit, offset int) ([]stock.Movement, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sm.id, sm.product_id, pr.product_code, pr.name,
		       sm.type, sm.quantity, sm.service_order_id,
		       u.username AS created_by_username, sm.created_by, sm.created_at,
		       COUNT(*) OVER () AS total
		FROM stock_movements sm
		JOIN products pr ON pr.id = sm.product_id
		JOIN users u ON u.id = sm.created_by
		WHERE (? = '' OR sm.product_id = ?)
		  AND (? = '' OR sm.type = ?)
		ORDER BY sm.created_at DESC
		LIMIT ? OFFSET ?
	`, productID, productID, movType, movType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movements []stock.Movement
	var total int64
	for rows.Next() {
		var m stock.Movement
		var id, productIDStr, createdByStr string
		var serviceOrderID sql.NullString
		var cat string
		var t int64
		if err := rows.Scan(&id, &productIDStr, &m.ProductCode, &m.ProductName,
			&m.Type, &m.Quantity, &serviceOrderID,
			&m.CreatedByUsername, &createdByStr, &cat, &t); err != nil {
			return nil, 0, err
		}
		m.ID, _ = uuid.Parse(id)
		m.ProductID, _ = uuid.Parse(productIDStr)
		m.CreatedBy, _ = uuid.Parse(createdByStr)
		m.CreatedAt = parseTime(cat)
		if serviceOrderID.Valid && serviceOrderID.String != "" {
			soID, _ := uuid.Parse(serviceOrderID.String)
			m.ServiceOrderID = &soID
		}
		total = t
		movements = append(movements, m)
	}
	return movements, total, rows.Err()
}

func (r *sqliteStockRepo) ApplyMovement(ctx context.Context, productID uuid.UUID, movType stock.MovementType, quantity int, userID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	recordedQty := quantity
	pidStr := productID.String()

	switch movType {
	case stock.MovementIn:
		res, err := tx.ExecContext(ctx, `
			UPDATE stock SET quantity = quantity + ?, updated_at = ? WHERE product_id = ?`,
			quantity, nowStr(), pidStr)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO stock (product_id, quantity) VALUES (?, ?)`, pidStr, quantity); err != nil {
				return err
			}
		}

	case stock.MovementOut:
		res, err := tx.ExecContext(ctx, `
			UPDATE stock SET quantity = quantity - ?, updated_at = ?
			WHERE product_id = ? AND quantity >= ?`,
			quantity, nowStr(), pidStr, quantity)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return stock.ErrInsufficientStock
		}

	case stock.MovementAdjustment:
		var current int
		err := tx.QueryRowContext(ctx, `SELECT quantity FROM stock WHERE product_id = ?`, pidStr).Scan(&current)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if quantity == current {
			return tx.Commit()
		}
		delta := quantity - current
		if delta < 0 {
			recordedQty = -delta
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO stock (product_id, quantity, updated_at) VALUES (?, ?, ?)
			ON CONFLICT (product_id) DO UPDATE SET quantity = excluded.quantity, updated_at = excluded.updated_at`,
			pidStr, quantity, nowStr()); err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_movements (product_id, type, quantity, created_by, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		pidStr, string(movType), recordedQty, userID.String(), nowStr())
	if err != nil {
		return err
	}

	return tx.Commit()
}
