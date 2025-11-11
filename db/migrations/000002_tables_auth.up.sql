CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    role role DEFAULT 'user'
 );

CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    id_users INT NOT NULL,
    fullname VARCHAR(50) DEFAULT 'user fullname',
    phoneNumber VARCHAR(12) DEFAULT '081234567890',
    address VARCHAR(100),
    photos VARCHAR(100),
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE account
ADD FOREIGN KEY (id_users) REFERENCES users(id);
