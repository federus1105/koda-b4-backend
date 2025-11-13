ALTER TABLE orders DROP CONSTRAINT "orders_id_delivery_fkey";
ALTER TABLE orders DROP CONSTRAINT "orders_id_account_fkey";
ALTER TABLE orders DROP CONSTRAINT "orders_id_paymentmethod_fkey";
ALTER TABLE orders DROP CONSTRAINT "orders_id_status_fkey";

DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS delivery;
DROP TABLE IF EXISTS status;