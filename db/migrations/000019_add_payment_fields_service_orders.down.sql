ALTER TABLE service_orders
  DROP COLUMN IF EXISTS pay_date,
  DROP COLUMN IF EXISTS is_paid;
