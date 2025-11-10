CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    id_account INT NOT NULL,
    id_paymentmethod INT NOT NULL,
    fullname VARCHAR(100) NOT NULL,
    address VARCHAR(100) NOT NULL,
    phoneNumber VARCHAR(100) NOT NULL,
    quantity FLOAT NOT NULL,
    delivery delivery NOT NULL,
    total FLOAT NOT NULL,
    status BOOLEAN NOT NULL,
    variant variant,
    size size,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE orders ADD FOREIGN KEY (id_account) REFERENCES account(id);

ALTER TABLE orders ADD FOREIGN KEY (id_paymentmethod) REFERENCES payment_method(id);