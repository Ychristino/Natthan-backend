package serviceOrder

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNotFound          = errors.New("ordem de serviço não encontrada")
	ErrInsufficientStock = errors.New("estoque insuficiente")
	ErrClientNotFound    = errors.New("cliente não encontrado")
	ErrProductNotFound   = errors.New("produto não encontrado")
	ErrServiceNotFound   = errors.New("serviço não encontrado")
)

func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

type Repository interface {
	CreateFull(ctx context.Context, so ServiceOrder, products []OrderProduct, services []OrderService, userID uuid.UUID) (*ServiceOrder, error)
	List(ctx context.Context, search, status, dateFrom, dateTo string, limit, offset int) ([]ServiceOrder, int64, error)
	GetBase(ctx context.Context, id uuid.UUID) (*ServiceOrder, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ServiceOrder, error)
	Update(ctx context.Context, so ServiceOrder) (*ServiceOrder, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	AddProduct(ctx context.Context, item OrderProduct, userID uuid.UUID) (*OrderProduct, error)
	RemoveProduct(ctx context.Context, orderID, itemID uuid.UUID) error
	AddService(ctx context.Context, item OrderService) (*OrderService, error)
	RemoveService(ctx context.Context, orderID, itemID uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const baseSelect = `
	SELECT so.id, so.client_id, so.employee_id, so.status, so.entry_date, so.exit_date,
	       so.pay_date, so.is_paid, so.notes, so.created_at, so.updated_at,
	       p.full_name  AS client_name,
	       ep.full_name AS employee_full_name
	FROM service_orders so
	JOIN persons p  ON p.id  = so.client_id
	JOIN persons ep ON ep.id = so.employee_id`

// ─── Tx helpers ───────────────────────────────────────────────────────────────

func deductStockInTx(ctx context.Context, tx pgx.Tx, productID uuid.UUID, qty int, serviceOrderID, userID uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE stock SET quantity = quantity - $1, updated_at = NOW()
		WHERE product_id = $2 AND quantity >= $1
	`, qty, pgtype.UUID{Bytes: productID, Valid: true})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInsufficientStock
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (product_id, type, quantity, service_order_id, created_by)
		VALUES ($1, 'out', $2, $3, $4)
	`,
		pgtype.UUID{Bytes: productID, Valid: true},
		qty,
		pgtype.UUID{Bytes: serviceOrderID, Valid: true},
		pgtype.UUID{Bytes: userID, Valid: true},
	)
	return err
}

func addOrderProductInTx(ctx context.Context, tx pgx.Tx, item OrderProduct) (*OrderProduct, error) {
	row := tx.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO service_order_products (id, service_order_id, product_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, product_id, quantity, unit_price, created_at
		)
		SELECT ins.id, ins.product_id, pr.product_code, pr.name, ins.quantity, ins.unit_price, ins.created_at
		FROM ins JOIN products pr ON pr.id = ins.product_id
	`,
		pgtype.UUID{Bytes: item.ID, Valid: true},
		pgtype.UUID{Bytes: item.ServiceOrderID, Valid: true},
		pgtype.UUID{Bytes: item.ProductID, Valid: true},
		item.Quantity, item.UnitPrice,
	)
	result, err := scanOrderProduct(row)
	if err != nil {
		if isFKViolation(err) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return result, nil
}

func addOrderServiceInTx(ctx context.Context, tx pgx.Tx, item OrderService) (*OrderService, error) {
	row := tx.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO service_order_services (id, service_order_id, service_id, price_applied, notes)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, service_id, price_applied, notes, created_at
		)
		SELECT ins.id, ins.service_id, sv.service_code, sv.name, ins.price_applied, ins.notes, ins.created_at
		FROM ins JOIN services sv ON sv.id = ins.service_id
	`,
		pgtype.UUID{Bytes: item.ID, Valid: true},
		pgtype.UUID{Bytes: item.ServiceOrderID, Valid: true},
		pgtype.UUID{Bytes: item.ServiceID, Valid: true},
		item.PriceApplied, item.Notes,
	)
	result, err := scanOrderService(row)
	if err != nil {
		if isFKViolation(err) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	return result, nil
}

// ─── Order CRUD ───────────────────────────────────────────────────────────────

// CreateFull inserts the order, all products (with stock deduction + movement),
// and all services in a single transaction. Any failure rolls back everything.
func (r *repository) CreateFull(ctx context.Context, so ServiceOrder, products []OrderProduct, services []OrderService, userID uuid.UUID) (*ServiceOrder, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var exitDate *pgtype.Timestamptz
	if so.ExitDate != nil {
		v := pgtype.Timestamptz{Time: *so.ExitDate, Valid: true}
		exitDate = &v
	}
	var payDate *pgtype.Timestamptz
	if so.PayDate != nil {
		v := pgtype.Timestamptz{Time: *so.PayDate, Valid: true}
		payDate = &v
	}

	var insertedID pgtype.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO service_orders (id, client_id, employee_id, status, entry_date, exit_date, pay_date, is_paid, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`,
		pgtype.UUID{Bytes: so.ID, Valid: true},
		pgtype.UUID{Bytes: so.ClientID, Valid: true},
		pgtype.UUID{Bytes: so.EmployeeID, Valid: true},
		so.Status, so.EntryDate, exitDate, payDate, so.IsPaid, so.Notes,
	).Scan(&insertedID)
	if err != nil {
		if isFKViolation(err) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	so.ID = uuid.UUID(insertedID.Bytes)

	// Fetch client/employee names within the same TX (own writes are visible).
	orderRow := tx.QueryRow(ctx, baseSelect+` WHERE so.id = $1`,
		pgtype.UUID{Bytes: so.ID, Valid: true})
	created, err := scanBase(orderRow)
	if err != nil {
		return nil, err
	}

	for i := range products {
		products[i].ServiceOrderID = created.ID
		if err := deductStockInTx(ctx, tx, products[i].ProductID, products[i].Quantity, created.ID, userID); err != nil {
			return nil, err
		}
		item, err := addOrderProductInTx(ctx, tx, products[i])
		if err != nil {
			return nil, err
		}
		created.Products = append(created.Products, *item)
	}

	for i := range services {
		services[i].ServiceOrderID = created.ID
		item, err := addOrderServiceInTx(ctx, tx, services[i])
		if err != nil {
			return nil, err
		}
		created.Services = append(created.Services, *item)
	}

	return created, tx.Commit(ctx)
}

func (r *repository) List(ctx context.Context, search, status, dateFrom, dateTo string, limit, offset int) ([]ServiceOrder, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT so.id, so.client_id, so.employee_id, so.status, so.entry_date, so.exit_date,
		       so.pay_date, so.is_paid, so.notes, so.created_at, so.updated_at,
		       p.full_name  AS client_name,
		       ep.full_name AS employee_full_name,
		       COUNT(*) OVER () AS total
		FROM service_orders so
		JOIN persons p  ON p.id  = so.client_id
		JOIN persons ep ON ep.id = so.employee_id
		WHERE ($1 = '' OR so.status::text = $1)
		  AND ($2 = '' OR p.full_name ILIKE '%' || $2 || '%' OR ep.full_name ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR so.entry_date::date >= $3::date)
		  AND ($4 = '' OR so.entry_date::date <= $4::date)
		ORDER BY so.entry_date DESC
		LIMIT $5 OFFSET $6
	`, status, search, dateFrom, dateTo, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []ServiceOrder
	var total int64
	for rows.Next() {
		so, t, err := scanWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		orders = append(orders, *so)
	}
	return orders, total, rows.Err()
}

func (r *repository) GetBase(ctx context.Context, id uuid.UUID) (*ServiceOrder, error) {
	row := r.db.QueryRow(ctx, baseSelect+` WHERE so.id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return scanBase(row)
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*ServiceOrder, error) {
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	var so *ServiceOrder
	var products []OrderProduct
	var services []OrderService

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var e error
		so, e = r.GetBase(gctx, id)
		return e
	})
	g.Go(func() error {
		var e error
		products, e = r.loadProducts(gctx, pgID)
		return e
	})
	g.Go(func() error {
		var e error
		services, e = r.loadServices(gctx, pgID)
		return e
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	so.Products = products
	so.Services = services
	return so, nil
}

func (r *repository) Update(ctx context.Context, so ServiceOrder) (*ServiceOrder, error) {
	var exitDate *pgtype.Timestamptz
	if so.ExitDate != nil {
		v := pgtype.Timestamptz{Time: *so.ExitDate, Valid: true}
		exitDate = &v
	}
	var payDate *pgtype.Timestamptz
	if so.PayDate != nil {
		v := pgtype.Timestamptz{Time: *so.PayDate, Valid: true}
		payDate = &v
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE service_orders
		SET status = $2, exit_date = $3, pay_date = $4, is_paid = $5, notes = $6, updated_at = NOW()
		WHERE id = $1
	`, pgtype.UUID{Bytes: so.ID, Valid: true}, so.Status, exitDate, payDate, so.IsPaid, so.Notes)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return r.GetBase(ctx, so.ID)
}

func (r *repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE service_orders SET status = $2, updated_at = NOW() WHERE id = $1
	`, pgtype.UUID{Bytes: id, Valid: true}, StatusCancelled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── Line items ───────────────────────────────────────────────────────────────

func (r *repository) AddProduct(ctx context.Context, item OrderProduct, userID uuid.UUID) (*OrderProduct, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := deductStockInTx(ctx, tx, item.ProductID, item.Quantity, item.ServiceOrderID, userID); err != nil {
		return nil, err
	}
	result, err := addOrderProductInTx(ctx, tx, item)
	if err != nil {
		return nil, err
	}
	return result, tx.Commit(ctx)
}

func (r *repository) RemoveProduct(ctx context.Context, orderID, itemID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM service_order_products WHERE id = $1 AND service_order_id = $2
	`, pgtype.UUID{Bytes: itemID, Valid: true}, pgtype.UUID{Bytes: orderID, Valid: true})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) AddService(ctx context.Context, item OrderService) (*OrderService, error) {
	row := r.db.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO service_order_services (id, service_order_id, service_id, price_applied, notes)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, service_id, price_applied, notes, created_at
		)
		SELECT ins.id, ins.service_id, sv.service_code, sv.name, ins.price_applied, ins.notes, ins.created_at
		FROM ins JOIN services sv ON sv.id = ins.service_id
	`,
		pgtype.UUID{Bytes: item.ID, Valid: true},
		pgtype.UUID{Bytes: item.ServiceOrderID, Valid: true},
		pgtype.UUID{Bytes: item.ServiceID, Valid: true},
		item.PriceApplied, item.Notes,
	)
	result, err := scanOrderService(row)
	if err != nil {
		if isFKViolation(err) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	return result, nil
}

func (r *repository) RemoveService(ctx context.Context, orderID, itemID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM service_order_services WHERE id = $1 AND service_order_id = $2
	`, pgtype.UUID{Bytes: itemID, Valid: true}, pgtype.UUID{Bytes: orderID, Valid: true})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── private loaders ──────────────────────────────────────────────────────────

func (r *repository) loadProducts(ctx context.Context, orderID pgtype.UUID) ([]OrderProduct, error) {
	rows, err := r.db.Query(ctx, `
		SELECT sop.id, sop.product_id, pr.product_code, pr.name, sop.quantity, sop.unit_price, sop.created_at
		FROM service_order_products sop
		JOIN products pr ON pr.id = sop.product_id
		WHERE sop.service_order_id = $1
		ORDER BY pr.name
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderProduct
	for rows.Next() {
		p, err := scanOrderProduct(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

func (r *repository) loadServices(ctx context.Context, orderID pgtype.UUID) ([]OrderService, error) {
	rows, err := r.db.Query(ctx, `
		SELECT sos.id, sos.service_id, sv.service_code, sv.name, sos.price_applied, sos.notes, sos.created_at
		FROM service_order_services sos
		JOIN services sv ON sv.id = sos.service_id
		WHERE sos.service_order_id = $1
		ORDER BY sv.name
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderService
	for rows.Next() {
		s, err := scanOrderService(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanBase(s scanner) (*ServiceOrder, error) {
	var so ServiceOrder
	var id, clientID, employeeID pgtype.UUID
	var exitDate, payDate pgtype.Timestamptz
	var notes pgtype.Text
	err := s.Scan(
		&id, &clientID, &employeeID, &so.Status, &so.EntryDate, &exitDate,
		&payDate, &so.IsPaid, &notes, &so.CreatedAt, &so.UpdatedAt,
		&so.ClientName, &so.EmployeeFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	so.ID = uuid.UUID(id.Bytes)
	so.ClientID = uuid.UUID(clientID.Bytes)
	so.EmployeeID = uuid.UUID(employeeID.Bytes)
	so.Notes = notes.String
	if exitDate.Valid {
		so.ExitDate = &exitDate.Time
	}
	if payDate.Valid {
		so.PayDate = &payDate.Time
	}
	return &so, nil
}

func scanWithTotal(s scanner) (*ServiceOrder, int64, error) {
	var so ServiceOrder
	var id, clientID, employeeID pgtype.UUID
	var exitDate, payDate pgtype.Timestamptz
	var notes pgtype.Text
	var total int64
	err := s.Scan(
		&id, &clientID, &employeeID, &so.Status, &so.EntryDate, &exitDate,
		&payDate, &so.IsPaid, &notes, &so.CreatedAt, &so.UpdatedAt,
		&so.ClientName, &so.EmployeeFullName, &total,
	)
	if err != nil {
		return nil, 0, err
	}
	so.ID = uuid.UUID(id.Bytes)
	so.ClientID = uuid.UUID(clientID.Bytes)
	so.EmployeeID = uuid.UUID(employeeID.Bytes)
	so.Notes = notes.String
	if exitDate.Valid {
		so.ExitDate = &exitDate.Time
	}
	if payDate.Valid {
		so.PayDate = &payDate.Time
	}
	return &so, total, nil
}

func scanOrderProduct(s scanner) (*OrderProduct, error) {
	var p OrderProduct
	var id, productID pgtype.UUID
	err := s.Scan(&id, &productID, &p.ProductCode, &p.Name, &p.Quantity, &p.UnitPrice, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	p.ID = uuid.UUID(id.Bytes)
	p.ProductID = uuid.UUID(productID.Bytes)
	return &p, nil
}

func scanOrderService(s scanner) (*OrderService, error) {
	var sv OrderService
	var id, serviceID pgtype.UUID
	var notes pgtype.Text
	err := s.Scan(&id, &serviceID, &sv.ServiceCode, &sv.Name, &sv.PriceApplied, &notes, &sv.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	sv.ID = uuid.UUID(id.Bytes)
	sv.ServiceID = uuid.UUID(serviceID.Bytes)
	sv.Notes = notes.String
	return &sv, nil
}
