CREATE TABLE product (
	id serial4 NOT NULL,
	name varchar(100) NOT NULL,
	id_product_images int4 NOT NULL,
	description varchar(255) NOT NULL,
	rating float8 NOT NULL,
	priceoriginal float8 NOT NULL,
	pricediscount float8 NULL,
	flash_sale bool NULL,
	stock int4 NOT NULL,
	createdat timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updatedat timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT product_pkey PRIMARY KEY (id),
	CONSTRAINT product_id_product_images_fkey FOREIGN KEY (id_product_images) REFERENCES product_images(id)
);