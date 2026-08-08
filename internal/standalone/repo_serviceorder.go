package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/serviceOrder"
)

type sqliteServiceOrderRepo struct{ db *sql.DB }

func NewServiceOrderRepository(db *sql.DB) serviceOrder.Repository {
	return &sqliteServiceOrderRepo{db: db}
}

const soBaseSelect = `
	SELECT so.id, so.client_id, so.employee_id, so.status, so.entry_date, so.exit_date,
	       so.pay_date, so.is_paid, so.notes, so.created_at, so.updated_at,
	       p.full_name  AS client_name,
	       ep.full_name AS employee_full_name
	FROM service_orders so
	JOIN persons p  ON p.id  = so.client_id
	JOIN persons ep ON ep.id = so.employee_id`

// ─── Tx helpers ───────────────────────────────────────────────────────────────

func deductStockInSQLiteTx(ctx context.Context, tx *sql.Tx, productID uuid.UUID, qty int, serviceOrderID, userID uuid.UUID) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE stock SET quantity = quantity - ?, updated_at = ?
		WHERE product_id = ? AND quantity >= ?`,
		qty, nowStr(), productID.String(), qty)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return serviceOrder.ErrInsufficientStock
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_movements (product_id, type, quantity, service_order_id, created_by, created_at)
		VALUES (?, 'out', ?, ?, ?, ?)`,
		productID.String(), qty, serviceOrderID.String(), userID.String(), nowStr())
	return err
}

func addOrderProductInSQLiteTx(ctx context.Context, tx *sql.Tx, item serviceOrder.OrderProduct) (*serviceOrder.OrderProduct, error) {
	now := nowStr()
	_, err := tx.ExecContext(ctx, `
		INSERT INTO service_order_products (id, service_order_id, product_id, quantity, unit_price, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		item.ID.String(), item.ServiceOrderID.String(), item.ProductID.String(),
		item.Quantity, item.UnitPrice, now)
	if err != nil {
		if isFKError(err) {
			return nil, serviceOrder.ErrProductNotFound
		}
		return nil, err
	}
	row := tx.QueryRowContext(ctx, `
		SELECT sop.id, sop.product_id, pr.product_code, pr.name, sop.quantity, sop.unit_price, sop.created_at
		FROM service_order_products sop
		JOIN products pr ON pr.id = sop.product_id
		WHERE sop.id = ?`, item.ID.String())
	return scanOrderProduct(row)
}

func addOrderServiceInSQLiteTx(ctx context.Context, tx *sql.Tx, item serviceOrder.OrderService) (*serviceOrder.OrderService, error) {
	now := nowStr()
	_, err := tx.ExecContext(ctx, `
		INSERT INTO service_order_services (id, service_order_id, service_id, price_applied, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		item.ID.String(), item.ServiceOrderID.String(), item.ServiceID.String(),
		item.PriceApplied, item.Notes, now)
	if err != nil {
		if isFKError(err) {
			return nil, serviceOrder.ErrServiceNotFound
		}
		return nil, err
	}
	row := tx.QueryRowContext(ctx, `
		SELECT sos.id, sos.service_id, sv.service_code, sv.name, sos.price_applied, sos.notes, sos.created_at
		FROM service_order_services sos
		JOIN services sv ON sv.id = sos.service_id
		WHERE sos.id = ?`, item.ID.String())
	return scanOrderService(row)
}

// ─── Order CRUD ───────────────────────────────────────────────────────────────

func (r *sqliteServiceOrderRepo) CreateFull(ctx context.Context, so serviceOrder.ServiceOrder, products []serviceOrder.OrderProduct, services []serviceOrder.OrderService, userID uuid.UUID) (*serviceOrder.ServiceOrder, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var exitDate sql.NullString
	if so.ExitDate != nil {
		exitDate = sql.NullString{String: so.ExitDate.UTC().Format("2006-01-02T15:04:05Z"), Valid: true}
	}
	var payDate sql.NullString
	if so.PayDate != nil {
		payDate = sql.NullString{String: so.PayDate.UTC().Format("2006-01-02T15:04:05Z"), Valid: true}
	}

	now := nowStr()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO service_orders (id, client_id, employee_id, status, entry_date, exit_date, pay_date, is_paid, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		so.ID.String(), so.ClientID.String(), so.EmployeeID.String(),
		string(so.Status), so.EntryDate.UTC().Format("2006-01-02T15:04:05Z"),
		exitDate, payDate, boolToInt(so.IsPaid), so.Notes, now, now)
	if err != nil {
		if isFKError(err) {
			return nil, serviceOrder.ErrClientNotFound
		}
		return nil, err
	}

	created, err := getSOBase(ctx, tx, so.ID.String())
	if err != nil {
		return nil, err
	}

	for i := range products {
		products[i].ServiceOrderID = created.ID
		if err := deductStockInSQLiteTx(ctx, tx, products[i].ProductID, products[i].Quantity, created.ID, userID); err != nil {
			return nil, err
		}
		item, err := addOrderProductInSQLiteTx(ctx, tx, products[i])
		if err != nil {
			return nil, err
		}
		created.Products = append(created.Products, *item)
	}

	for i := range services {
		services[i].ServiceOrderID = created.ID
		item, err := addOrderServiceInSQLiteTx(ctx, tx, services[i])
		if err != nil {
			return nil, err
		}
		created.Services = append(created.Services, *item)
	}

	return created, tx.Commit()
}

func (r *sqliteServiceOrderRepo) List(ctx context.Context, search, status, dateFrom, dateTo string, limit, offset int) ([]serviceOrder.ServiceOrder, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT so.id, so.client_id, so.employee_id, so.status, so.entry_date, so.exit_date,
		       so.pay_date, so.is_paid, so.notes, so.created_at, so.updated_at,
		       p.full_name  AS client_name,
		       ep.full_name AS employee_full_name,
		       COUNT(*) OVER () AS total
		FROM service_orders so
		JOIN persons p  ON p.id  = so.client_id
		JOIN persons ep ON ep.id = so.employee_id
		WHERE (? = '' OR so.status = ?)
		  AND (? = '' OR LOWER(p.full_name) LIKE LOWER('%' || ? || '%') OR LOWER(ep.full_name) LIKE LOWER('%' || ? || '%'))
		  AND (? = '' OR DATE(so.entry_date) >= DATE(?))
		  AND (? = '' OR DATE(so.entry_date) <= DATE(?))
		ORDER BY so.entry_date DESC
		LIMIT ? OFFSET ?
	`, status, status, search, search, search, dateFrom, dateFrom, dateTo, dateTo, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []serviceOrder.ServiceOrder
	var total int64
	for rows.Next() {
		so, t, err := scanSOWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		orders = append(orders, *so)
	}
	return orders, total, rows.Err()
}

func (r *sqliteServiceOrderRepo) GetBase(ctx context.Context, id uuid.UUID) (*serviceOrder.ServiceOrder, error) {
	row := r.db.QueryRowContext(ctx, soBaseSelect+` WHERE so.id = ?`, id.String())
	return scanSOBase(row)
}

func (r *sqliteServiceOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*serviceOrder.ServiceOrder, error) {
	so, err := r.GetBase(ctx, id)
	if err != nil {
		return nil, err
	}
	so.Products, err = r.loadProducts(ctx, id.String())
	if err != nil {
		return nil, err
	}
	so.Services, err = r.loadServices(ctx, id.String())
	if err != nil {
		return nil, err
	}
	return so, nil
}

func (r *sqliteServiceOrderRepo) Update(ctx context.Context, so serviceOrder.ServiceOrder) (*serviceOrder.ServiceOrder, error) {
	var exitDate sql.NullString
	if so.ExitDate != nil {
		exitDate = sql.NullString{String: so.ExitDate.UTC().Format("2006-01-02T15:04:05Z"), Valid: true}
	}
	var payDate sql.NullString
	if so.PayDate != nil {
		payDate = sql.NullString{String: so.PayDate.UTC().Format("2006-01-02T15:04:05Z"), Valid: true}
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE service_orders
		SET status = ?, exit_date = ?, pay_date = ?, is_paid = ?, notes = ?, updated_at = ?
		WHERE id = ?`,
		string(so.Status), exitDate, payDate, boolToInt(so.IsPaid), so.Notes, nowStr(), so.ID.String())
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, serviceOrder.ErrNotFound
	}
	return r.GetBase(ctx, so.ID)
}

func (r *sqliteServiceOrderRepo) Deactivate(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE service_orders SET status = ?, updated_at = ? WHERE id = ?`,
		string(serviceOrder.StatusCancelled), nowStr(), id.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return serviceOrder.ErrNotFound
	}
	return nil
}

// ─── Line items ───────────────────────────────────────────────────────────────

func (r *sqliteServiceOrderRepo) AddProduct(ctx context.Context, item serviceOrder.OrderProduct, userID uuid.UUID) (*serviceOrder.OrderProduct, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := deductStockInSQLiteTx(ctx, tx, item.ProductID, item.Quantity, item.ServiceOrderID, userID); err != nil {
		return nil, err
	}
	result, err := addOrderProductInSQLiteTx(ctx, tx, item)
	if err != nil {
		return nil, err
	}
	return result, tx.Commit()
}

func (r *sqliteServiceOrderRepo) RemoveProduct(ctx context.Context, orderID, itemID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM service_order_products WHERE id = ? AND service_order_id = ?`,
		itemID.String(), orderID.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return serviceOrder.ErrNotFound
	}
	return nil
}

func (r *sqliteServiceOrderRepo) AddService(ctx context.Context, item serviceOrder.OrderService) (*serviceOrder.OrderService, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO service_order_services (id, service_order_id, service_id, price_applied, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		item.ID.String(), item.ServiceOrderID.String(), item.ServiceID.String(),
		item.PriceApplied, item.Notes, now)
	if err != nil {
		if isFKError(err) {
			return nil, serviceOrder.ErrServiceNotFound
		}
		return nil, err
	}
	row := r.db.QueryRowContext(ctx, `
		SELECT sos.id, sos.service_id, sv.service_code, sv.name, sos.price_applied, sos.notes, sos.created_at
		FROM service_order_services sos
		JOIN services sv ON sv.id = sos.service_id
		WHERE sos.id = ?`, item.ID.String())
	return scanOrderService(row)
}

func (r *sqliteServiceOrderRepo) RemoveService(ctx context.Context, orderID, itemID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM service_order_services WHERE id = ? AND service_order_id = ?`,
		itemID.String(), orderID.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return serviceOrder.ErrNotFound
	}
	return nil
}

// ─── private loaders ──────────────────────────────────────────────────────────

func (r *sqliteServiceOrderRepo) loadProducts(ctx context.Context, orderID string) ([]serviceOrder.OrderProduct, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sop.id, sop.product_id, pr.product_code, pr.name, sop.quantity, sop.unit_price, sop.created_at
		FROM service_order_products sop
		JOIN products pr ON pr.id = sop.product_id
		WHERE sop.service_order_id = ?
		ORDER BY pr.name
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []serviceOrder.OrderProduct
	for rows.Next() {
		p, err := scanOrderProduct(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

func (r *sqliteServiceOrderRepo) loadServices(ctx context.Context, orderID string) ([]serviceOrder.OrderService, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sos.id, sos.service_id, sv.service_code, sv.name, sos.price_applied, sos.notes, sos.created_at
		FROM service_order_services sos
		JOIN services sv ON sv.id = sos.service_id
		WHERE sos.service_order_id = ?
		ORDER BY sv.name
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []serviceOrder.OrderService
	for rows.Next() {
		s, err := scanOrderService(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

// ─── tx helper for CreateFull ─────────────────────────────────────────────────

func getSOBase(ctx context.Context, tx *sql.Tx, id string) (*serviceOrder.ServiceOrder, error) {
	row := tx.QueryRowContext(ctx, soBaseSelect+` WHERE so.id = ?`, id)
	return scanSOBase(row)
}

// ─── scanners ─────────────────────────────────────────────────────────────────

func scanSOBase(s rowScanner) (*serviceOrder.ServiceOrder, error) {
	var so serviceOrder.ServiceOrder
	var id, clientID, employeeID, statusStr, entryDate string
	var exitDate, payDate, notes sql.NullString
	var isPaidVal int
	var cat, uat string
	err := s.Scan(
		&id, &clientID, &employeeID, &statusStr, &entryDate,
		&exitDate, &payDate, &isPaidVal, &notes,
		&cat, &uat,
		&so.ClientName, &so.EmployeeFullName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, serviceOrder.ErrNotFound
		}
		return nil, err
	}
	so.ID, _ = uuid.Parse(id)
	so.ClientID, _ = uuid.Parse(clientID)
	so.EmployeeID, _ = uuid.Parse(employeeID)
	so.Status = serviceOrder.Status(statusStr)
	so.EntryDate = parseTime(entryDate)
	so.IsPaid = isPaidVal == 1
	so.Notes = notes.String
	so.CreatedAt = parseTime(cat)
	so.UpdatedAt = parseTime(uat)
	if exitDate.Valid && exitDate.String != "" {
		t := parseTime(exitDate.String)
		so.ExitDate = &t
	}
	if payDate.Valid && payDate.String != "" {
		t := parseTime(payDate.String)
		so.PayDate = &t
	}
	return &so, nil
}

func scanSOWithTotal(s rowScanner) (*serviceOrder.ServiceOrder, int64, error) {
	var so serviceOrder.ServiceOrder
	var id, clientID, employeeID, statusStr, entryDate string
	var exitDate, payDate, notes sql.NullString
	var isPaidVal int
	var cat, uat string
	var total int64
	err := s.Scan(
		&id, &clientID, &employeeID, &statusStr, &entryDate,
		&exitDate, &payDate, &isPaidVal, &notes,
		&cat, &uat,
		&so.ClientName, &so.EmployeeFullName, &total,
	)
	if err != nil {
		return nil, 0, err
	}
	so.ID, _ = uuid.Parse(id)
	so.ClientID, _ = uuid.Parse(clientID)
	so.EmployeeID, _ = uuid.Parse(employeeID)
	so.Status = serviceOrder.Status(statusStr)
	so.EntryDate = parseTime(entryDate)
	so.IsPaid = isPaidVal == 1
	so.Notes = notes.String
	so.CreatedAt = parseTime(cat)
	so.UpdatedAt = parseTime(uat)
	if exitDate.Valid && exitDate.String != "" {
		t := parseTime(exitDate.String)
		so.ExitDate = &t
	}
	if payDate.Valid && payDate.String != "" {
		t := parseTime(payDate.String)
		so.PayDate = &t
	}
	return &so, total, nil
}

func scanOrderProduct(s rowScanner) (*serviceOrder.OrderProduct, error) {
	var p serviceOrder.OrderProduct
	var id, productID, cat string
	err := s.Scan(&id, &productID, &p.ProductCode, &p.Name, &p.Quantity, &p.UnitPrice, &cat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, serviceOrder.ErrNotFound
		}
		return nil, err
	}
	p.ID, _ = uuid.Parse(id)
	p.ProductID, _ = uuid.Parse(productID)
	p.CreatedAt = parseTime(cat)
	return &p, nil
}

func scanOrderService(s rowScanner) (*serviceOrder.OrderService, error) {
	var sv serviceOrder.OrderService
	var id, serviceID, cat string
	var notes sql.NullString
	err := s.Scan(&id, &serviceID, &sv.ServiceCode, &sv.Name, &sv.PriceApplied, &notes, &cat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, serviceOrder.ErrNotFound
		}
		return nil, err
	}
	sv.ID, _ = uuid.Parse(id)
	sv.ServiceID, _ = uuid.Parse(serviceID)
	sv.Notes = notes.String
	sv.CreatedAt = parseTime(cat)
	return &sv, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
