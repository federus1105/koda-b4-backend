CREATE TABLE delivery (
    id serial PRIMARY KEY,
    name varchar(100) NOT NULL
);


CREATE TABLE status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL DEFAULT 'on progres'
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    id_account INT NOT NULL,
    id_paymentmethod INT NOT NULL,
    fullname VARCHAR(100) NOT NULL,
    address VARCHAR(100) NOT NULL,
    phoneNumber VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    id_delivery INT NOT NULL,
    total FLOAT NOT NULL,
    id_status INT NOT NULL,
    order_number VARCHAR(50) NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE orders ADD FOREIGN KEY(id_delivery) REFERENCES delivery(id);

ALTER TABLE orders ADD FOREIGN KEY (id_account) REFERENCES account(id);

ALTER TABLE orders ADD FOREIGN KEY (id_paymentmethod) REFERENCES payment_method(id);

ALTER TABLE orders ADD FOREIGN KEY(id_status) REFERENCES status(id);