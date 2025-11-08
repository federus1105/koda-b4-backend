CREATE TABLE payment_method (
	id serial4 NOT NULL,
	name varchar(50) NOT NULL,
	photos varchar(255) NOT NULL,
	CONSTRAINT payment_method_pkey PRIMARY KEY (id)
);