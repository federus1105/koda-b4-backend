CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE product_images(
    id SERIAL PRIMARY KEY,
    photos_one VARCHAR(255) NOT NULL,
    photos_two VARCHAR(255),
    photos_three VARCHAR(255),
    photos_four VARCHAR(255)
);


CREATE TABLE product (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    id_product_images INT NOT NULL,
    description VARCHAR(255) NOT NULL,
    RATING FLOAT NOT NULL,
    priceOriginal FLOAT NOT NULL,
    priceDiscount FLOAT,
    flash_sale BOOLEAN,
    stock INT NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_orders (
    id_product INT NOT NULL,
    id_order INT NOT NULL,
    FOREIGN KEY (id_product) REFERENCES product(id),
    FOREIGN KEY (id_order) REFERENCES orders(id)
);


CREATE TABLE product_categories (
    id_product INT NOT NULL,
    id_categories INT NOT NULL,
    FOREIGN KEY (id_product) REFERENCES product(id),
    FOREIGN KEY (id_categories) REFERENCES categories(id)
);

ALTER TABLE product ADD FOREIGN KEY (id_product_images) REFERENCES product_images(id);
ALTER TABLE product_orders ADD FOREIGN KEY (id_order) REFERENCES orders(id);
ALTER TABLE product_orders ADD FOREIGN KEY (id_product) REFERENCES product(id);
ALTER TABLE product_categories ADD FOREIGN KEY (id_categories) REFERENCES categories(id);
ALTER TABLE product_categories ADD FOREIGN KEY (id_product) REFERENCES product(id)

