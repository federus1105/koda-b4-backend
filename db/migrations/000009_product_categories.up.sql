CREATE TABLE product_categories (
	id_product int4 NOT NULL,
	id_categories int4 NOT NULL,
	CONSTRAINT product_categories_id_categories_fkey FOREIGN KEY (id_categories) REFERENCES categories(id),
	CONSTRAINT product_categories_id_product_fkey FOREIGN KEY (id_product) REFERENCES product(id)
);