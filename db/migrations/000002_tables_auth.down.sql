ALTER TABLE account DROP CONSTRAINT "account_id_users_fkey";

DROP TABLE IF EXISTS account;

DROP TABLE IF EXISTS users;