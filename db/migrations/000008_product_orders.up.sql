CREATE TABLE product_orders (
	id_product int4 NOT NULL,
	id_order int4 NOT NULL,
	CONSTRAINT product_orders_id_order_fkey FOREIGN KEY (id_order) REFERENCES orders(id),
	CONSTRAINT product_orders_id_product_fkey FOREIGN KEY (id_product) REFERENCES product(id)
);