CREATE TABLE account (
	id serial4 NOT NULL,
	id_users int4 NOT NULL,
	fullname varchar(50) DEFAULT 'user'::character varying NULL,
	phonenumber varchar(12) DEFAULT '+62812342'::character varying NULL,
	address varchar(100) NOT NULL,
	photos varchar(100) NOT NULL,
	createdat timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updatedat timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT account_pkey PRIMARY KEY (id),
	CONSTRAINT account_id_users_fkey FOREIGN KEY (id_users) REFERENCES users(id)
);