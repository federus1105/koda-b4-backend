ALTER TABLE orders DROP CONSTRAINT "orders_id_account_fkey";
ALTER TABLE orders DROP CONSTRAINT "orders_id_paymentmethod_fkey";
DROP TABLE IF EXISTS orders;