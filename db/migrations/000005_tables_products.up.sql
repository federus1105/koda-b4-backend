CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    photos_one VARCHAR(255),
    photos_two VARCHAR(255),
    photos_three VARCHAR(255),
    photos_four VARCHAR(255)
);

CREATE TABLE sizes (
    id SERIAL PRIMARY KEY,
    name size NOT NULL
);

CREATE TABLE variants (
    id SERIAL PRIMARY KEY,
    name variant not NULL
);

CREATE TABLE product (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    id_product_images INT NOT NULL,
    description VARCHAR(255) NOT NULL,
    RATING FLOAT NOT NULL,
    priceOriginal FLOAT NOT NULL,
    priceDiscount FLOAT DEFAULT 0,
    flash_sale BOOLEAN DEFAULT FALSE,
    stock INT NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    is_favorite BOOLEAN DEFAULT FALSE,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE variant_product (
    id_product INT NOT NULL,
    id_variant int NOT NULL
);

CREATE TABLE size_product (
    id_product INT NOT NULL,
    id_size int NOT NULL
);

CREATE TABLE product_orders (
    id_product INT NOT NULL,
    id_order INT NOT NULL,
    quantity INT NOT NULL,
    subtotal FLOAT NOT NULL,
    size VARCHAR(30),
    variant VARCHAR(30),
    FOREIGN KEY (id_product) REFERENCES product(id),
    FOREIGN KEY (id_order) REFERENCES orders(id)
);


CREATE TABLE product_categories (
    id_product INT NOT NULL,
    id_categories INT NOT NULL,
    FOREIGN KEY (id_product) REFERENCES product(id),
    FOREIGN KEY (id_categories) REFERENCES categories(id)
);

CREATE TABLE cart (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    product_id INT NOT NULL,
    size_id INT REFERENCES sizes(id),
    variant_id INT REFERENCES variants(id),
    quantity INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE cart ADD CONSTRAINT unique_cart_item UNIQUE (account_id, product_id, size_id, variant_id);
ALTER TABLE cart ADD FOREIGN KEY (account_id) REFERENCES account(id);
ALTER TABLE cart ADD FOREIGN KEY (product_id)  REFERENCES product(id);
ALTER TABLE size_product ADD FOREIGN KEY (id_size) REFERENCES sizes(id);
ALTER TABLE size_product ADD FOREIGN KEY (id_product) REFERENCES product(id);
ALTER TABLE variant_product ADD FOREIGN KEY (id_variant) REFERENCES variants(id);
ALTER TABLE variant_product ADD FOREIGN KEY (id_product) REFERENCES product(id);
ALTER TABLE product ADD FOREIGN KEY (id_product_images) REFERENCES product_images(id);
ALTER TABLE product_orders ADD FOREIGN KEY (id_order) REFERENCES orders(id);
ALTER TABLE product_orders ADD FOREIGN KEY (id_product) REFERENCES product(id);
ALTER TABLE product_categories ADD FOREIGN KEY (id_categories) REFERENCES categories(id);
ALTER TABLE product_categories ADD FOREIGN KEY (id_product) REFERENCES product(id);

