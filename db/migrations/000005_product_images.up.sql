CREATE TABLE product_images (
	id serial4 NOT NULL,
	photos_one varchar(255) NOT NULL,
	photos_two varchar(255) NULL,
	photos_three varchar(255) NULL,
	photos_four varchar(255) NULL,
	CONSTRAINT product_images_pkey PRIMARY KEY (id)
);