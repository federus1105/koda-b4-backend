#  ☕ Coffee Shop Admin Backend API
> Coffee Shop Admin Backend is a backend API designed to manage coffee shop administration systems from product management, orders, users, to uploading product images in multiple files. Built with an efficient, secure, and easily scalable architecture to meet the needs of a growing coffee business.

 
## 📸 Preview
### Swagger Documentation
![alt text](/db//erd//image.png)
### Table ERD Coffe-shop
```mermaid
erDiagram
ROLE {
    ENUM admin
    ENUM user
    }

SIZE {
    ENUM regular
    ENUM medium
    ENUM large
}

VARIANT {
    ENUM ice
    ENUM hot
}

USERS {
    int id
    string email
    string password
    ROLE role
    }

ACCOUNT {
    int id
    int id_users
    string fullname
    string phoneNumber
    string address
    string ptotos
    timestamp createdAt
    timestamp updatedAt
}

VARIANT_PRODUCT {
    int id
    VARIANT name 
}

SIZE_PRODUCT {
    int id
    SIZE name
}

ORDERS {
    int id
    int id_account
    int id_paymenMethod
    string fullname
    string address
    string phoneNumber
    DELIVERY delivery
    float total
    boolean status 
    timestamp createdAt
}

PAYMENT_METHOD {
    int id
    string name
    string photos
}

PRODUCT_IMAGES {
    int id
    string photos_one
    string photos_two
    string photos_three
    string photos_four
    timestamp createdAt
    timestamp updatedAt
}

PRODUCT {
    int id
    string name
    string description
    int id_product_images
    float rating
    float priceOriginal
    float priceDiscount
    boolean flash_sale
    int stock
    boolen is_deleted
    boolen is_favorite
    timestamp createdAt
    timestamp updatedAt
}

CATEGORIES {
    int id
    string name
}

PRODUCT_CATEGORIES {
    int id_product
    int id_categories
}

PRODUCT_ORDERS {
   int id_product
   int id_order
   int quantity
   float subtotal
}


    ROLE ||--o{ USERS : ""
    USERS ||--|| ACCOUNT : ""
    CATEGORIES||--o{ PRODUCT_CATEGORIES : ""

    PRODUCT ||--o{PRODUCT_CATEGORIES :""
    
    PAYMENT_METHOD  ||--||ORDERS:""
    ACCOUNT ||--o{ORDERS :""

    ORDERS ||--o{PRODUCT_ORDERS: ""
    PRODUCT ||--o{PRODUCT_ORDERS :""

    PRODUCT_IMAGES |o--|{PRODUCT: ""

    ORDERS ||--o{SIZE_PRODUCT: ""
    ORDERS ||--o{VARIANT_PRODUCT: ""

    SIZE_PRODUCT ||--|{ SIZE:""
    VARIANT_PRODUCT ||--o{VARIANT:""

```

## Redis Cache Overview ⚡
| Status                 | Description                                                                                        | Response Time | Screenshot                                      |
| ---------------------- | ------------------------------------------------------------------------------------- | ------------ | ----------------------------------------------- |
| **Before Using Cache** | Data is still taken directly from the database, so it takes quite a long time. | ⏳ Slow          | ![alt text](</docs/images/before.png>) |
| **After Using Cache**  | Data is taken from Redis Cache so the process becomes faster.                  | ⚡ Fast     | ![alt text](</docs/images/after.png>) |

<br>

## 🚀 Features
- 🔐 JWT Authentication (Login & Register)
- ✨ Multiple File Upload
- 📘 Swagger Auto-Generated API Docs
- 🧾 CRUD for resources
- 📦 Manajemen Products, Orders & Users
- 🗂️ MVC architecture
- 📦 PostgreSQL integration
- 👤 Autentikasi & Otorisasi Admin


## 🛠️ Tech Stack
![Go](https://img.shields.io/badge/-Go-00ADD8?logo=go&logoColor=white&style=for-the-badge)
![Gin](https://img.shields.io/badge/-Gin-00ADD8?logo=go&logoColor=white&style=for-the-badge)
![PostgreSQL](https://img.shields.io/badge/-PostgreSQL-4169E1?logo=postgresql&logoColor=white&style=for-the-badge)
![Swagger](https://img.shields.io/badge/Swagger-UI-85EA2D?logo=swagger&logoColor=black&style=for-the-badge)

##  🔐 .env Configuration
```
DBUSER=youruser
DBPASS=yourpass
DBHOST=localhost
DBPORT=5432
DBNAME=tickitz

JWT_SECRET=your_jwt_secret
```

## 📦 How to Install & Run Project
### First, clone this repository: 
```
https://github.com/federus1105/koda-b4-backend.git
```
### Install Dependencies
```go
go mod tidy
```
### Run Server/Project
```go
go run .\cmd\main.go 
```
### Init Swagger
```go
swag init -g ./cmd/main.go
```
### Open Swagger Documentation in Browser
#### ⚠️ Make sure the server is running
```http://localhost:8011/swagger/index.html```


<br>


## 🗃️ How to run Database Migrations
### ⚠️ Attention: This only applies to PostgreSQL, because enums can only be used in PostgreSQL.
### Install Go migrate
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest;
```
### Create database
```bash
CREATE DATABASE <database_name>;
```
### Migrations Up
```bash
migrate -path ./db/migrations -database "postgres://user:password@localhost:5432/database?sslmode=disable" up
```
### Migrations Down
```bash
migrate -path ./db/migrations -database "postgres://user:password@localhost:5432/database?sslmode=disable" down
```

## 👨‍💻 Made with by
📫 [federusrudi@gmail.com](mailto:federusrudi@gmail.com)  
💼 [LinkedIn](https://www.linkedin.com/in/federus-rudi/)  

## 📜 License
Released under the **MIT License**.  
You’re free to use, modify, and distribute this project — just don’t forget to give a little credit

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

