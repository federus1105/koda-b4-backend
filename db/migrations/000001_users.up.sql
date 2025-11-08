CREATE TYPE role AS ENUM ('admin', 'user');

CREATE TABLE users (
	id serial4 NOT NULL,
	email varchar(100) NOT NULL,
	password varchar(100) NOT NULL,
	role role NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id)
);