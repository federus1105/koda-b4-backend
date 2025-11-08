CREATE TYPE size AS ENUM ('regular', 'medium', 'large');
CREATE TYPE variant AS ENUM ('ice', 'hot');
CREATE TYPE delivery AS ENUM ('dineIn','doorDelivery', 'pickUp');

CREATE TABLE orders (
	id serial4 NOT NULL,
	id_account int4 NOT NULL,
	id_paymentmethod int4 NOT NULL,
	fullname varchar(100) NOT NULL,
	address varchar(100) NOT NULL,
	phonenumber varchar(100) NOT NULL,
	quantity FLOAT NOT NULL,
	delivery delivery NOT NULL,
	total float8 NOT NULL,
	status bool NOT NULL,
	createdat timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	size size NULL,
	variant variant NULL,
	CONSTRAINT orders_pkey PRIMARY KEY (id),
	CONSTRAINT orders_id_account_fkey FOREIGN KEY (id_account) REFERENCES account(id),
	CONSTRAINT orders_id_paymentmethod_fkey FOREIGN KEY (id_paymentmethod) REFERENCES payment_method(id)
);