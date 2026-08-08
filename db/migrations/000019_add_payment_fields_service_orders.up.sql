ALTER TABLE service_orders
  ADD COLUMN pay_date TIMESTAMPTZ,
  ADD COLUMN is_paid  BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_service_orders_pay_date_unpaid
  ON service_orders (pay_date)
  WHERE is_paid = FALSE AND pay_date IS NOT NULL;
