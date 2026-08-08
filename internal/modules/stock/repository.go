package stock

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientStock = errors.New("estoque insuficiente")

type Repository interface {
	List(ctx context.Context, name string, limit, offset int) ([]StockItem, int64, error)
	ListMovements(ctx context.Context, productID, movType string, limit, offset int) ([]Movement, int64, error)
	ApplyMovement(ctx context.Context, productID uuid.UUID, movType MovementType, quantity int, userID uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func (r *repository) List(ctx context.Context, name string, limit, offset int) ([]StockItem, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.product_id, pr.product_code, pr.name, s.quantity,
		       MAX(sm.created_at) AS last_movement_at,
		       s.updated_at,
		       COUNT(*) OVER () AS total
		FROM stock s
		JOIN products pr ON pr.id = s.product_id
		LEFT JOIN stock_movements sm ON sm.product_id = s.product_id
		WHERE ($1 = '' OR pr.name ILIKE '%' || $1 || '%' OR pr.product_code ILIKE '%' || $1 || '%')
		GROUP BY s.id, s.product_id, pr.product_code, pr.name, s.quantity, s.updated_at
		ORDER BY pr.name
		LIMIT $2 OFFSET $3
	`, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []StockItem
	var total int64
	for rows.Next() {
		var item StockItem
		var id, productID pgtype.UUID
		var lastMovAt pgtype.Timestamptz
		err := rows.Scan(&id, &productID, &item.ProductCode, &item.ProductName, &item.Quantity,
			&lastMovAt, &item.UpdatedAt, &total)
		if err != nil {
			return nil, 0, err
		}
		item.ID = uuid.UUID(id.Bytes)
		item.ProductID = uuid.UUID(productID.Bytes)
		if lastMovAt.Valid {
			item.LastMovementAt = &lastMovAt.Time
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *repository) ListMovements(ctx context.Context, productID, movType string, limit, offset int) ([]Movement, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT sm.id, sm.product_id, pr.product_code, pr.name,
		       sm.type, sm.quantity, sm.service_order_id,
		       u.username AS created_by_username, sm.created_at,
		       COUNT(*) OVER () AS total
		FROM stock_movements sm
		JOIN products pr ON pr.id = sm.product_id
		JOIN users u ON u.id = sm.created_by
		WHERE ($1 = '' OR sm.product_id::text = $1)
		  AND ($2 = '' OR sm.type::text = $2)
		ORDER BY sm.created_at DESC
		LIMIT $3 OFFSET $4
	`, productID, movType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movements []Movement
	var total int64
	for rows.Next() {
		var m Movement
		var id, productID, createdBy pgtype.UUID
		var serviceOrderID pgtype.UUID
		err := rows.Scan(&id, &productID, &m.ProductCode, &m.ProductName,
			&m.Type, &m.Quantity, &serviceOrderID,
			&m.CreatedByUsername, &m.CreatedAt, &total)
		if err != nil {
			return nil, 0, err
		}
		m.ID = uuid.UUID(id.Bytes)
		m.ProductID = uuid.UUID(productID.Bytes)
		m.CreatedBy = uuid.UUID(createdBy.Bytes)
		if serviceOrderID.Valid {
			id := uuid.UUID(serviceOrderID.Bytes)
			m.ServiceOrderID = &id
		}
		movements = append(movements, m)
	}
	return movements, total, rows.Err()
}

func (r *repository) ApplyMovement(ctx context.Context, productID uuid.UUID, movType MovementType, quantity int, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	recordedQty := quantity

	switch movType {
	case MovementIn:
		tag, err := tx.Exec(ctx, `
			UPDATE stock SET quantity = quantity + $1, updated_at = NOW()
			WHERE product_id = $2
		`, quantity, pgUUID(productID))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			// First time this product appears in stock — create the row.
			if _, err := tx.Exec(ctx, `
				INSERT INTO stock (product_id, quantity) VALUES ($1, $2)
			`, pgUUID(productID), quantity); err != nil {
				return err
			}
		}

	case MovementOut:
		tag, err := tx.Exec(ctx, `
			UPDATE stock SET quantity = quantity - $1, updated_at = NOW()
			WHERE product_id = $2 AND quantity >= $1
		`, quantity, pgUUID(productID))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrInsufficientStock
		}

	case MovementAdjustment:
		var current int
		err := tx.QueryRow(ctx, `SELECT quantity FROM stock WHERE product_id = $1`, pgUUID(productID)).Scan(&current)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if quantity == current {
			// Nothing to change.
			return tx.Commit(ctx)
		}
		delta := quantity - current
		if delta < 0 {
			recordedQty = -delta
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO stock (product_id, quantity) VALUES ($1, $2)
			ON CONFLICT (product_id) DO UPDATE SET quantity = $2, updated_at = NOW()
		`, pgUUID(productID), quantity); err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (product_id, type, quantity, created_by)
		VALUES ($1, $2, $3, $4)
	`, pgUUID(productID), movType, recordedQty, pgUUID(userID))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
