#  ☕ Coffee Shop Senja Kopi Kiri
> The Senja Kopi Kiri app allows users to order coffee products directly through the app, from adding products to their cart, making payments using available payment methods, to viewing their order history and details. Users can also update information on their profile page for a more personalized and flexible experience. Furthermore, the app supports a forgot password feature, where users can request a password reset and receive a token via their registered email. The system has two roles: user and admin, with the admin responsible for managing the categories and products available in the app.

 
## 📸 Preview
### Swagger Documentation
![alt text](/assets/swagger.png)
### Table ERD Coffe-shop Senja Kopi Kiri
```mermaid
erDiagram
ROLE {
    ENUM admin
    ENUM user
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


VARIANT_ENUM {
    ENUM ice
    ENUM hot
}

VARIANT {
    int id
    VARIANT name 
}

PRODUCT_VARIANT {
    int id_variant_product
    int id_product
}

SIZE_ENUM {
    ENUM regular
    ENUM medium
    ENUM large
}

SIZE {
    int id
    SIZE name
}

PRODUCT_SIZE {
    int id_size_product
    int id_product
}

DELIVERY{
    int id
    string name
}

ORDERS {
    int id
    int id_account
    int id_paymenMethod
    string fullname
    string email
    string address
    string phoneNumber
    int id_delivery
    float total
    int IDstatus 
    int order_number
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
    int id_size
    int id_variant
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
   string variant
   string size
   float subtotal
}

CART {
    int id
    int account_id
    int product_id
    int size_id
    int variant_id
    float quantity
    timestamp created_at
    timestamp updated_at
}

STATUS {
    int id
    string name 
}


    ROLE ||--o{ USERS : ""
    USERS ||--|| ACCOUNT : ""
    CATEGORIES||--o{ PRODUCT_CATEGORIES : ""

    PRODUCT_CATEGORIES ||--o{PRODUCT :""
    
    PAYMENT_METHOD  ||--||ORDERS:""
    ACCOUNT ||--o{ORDERS :""

    ORDERS ||--o{PRODUCT_ORDERS: ""
    PRODUCT ||--o{PRODUCT_ORDERS :""

    PRODUCT_IMAGES |o--|{PRODUCT: ""

    SIZE ||--o{SIZE_ENUM:""
    PRODUCT_SIZE ||--o{SIZE:""

    VARIANT ||--o{VARIANT_ENUM:""
    PRODUCT_VARIANT||--o{VARIANT:""

    PRODUCT ||--o{PRODUCT_SIZE:""
    PRODUCT ||--o{PRODUCT_VARIANT:""

    DELIVERY ||--||ORDERS:""

    CART||--o{ACCOUNT:""
    PRODUCT||--||CART:""

    STATUS ||--||ORDERS:""


```

## Redis Cache Overview ⚡
| Status                 | Description                                                                                        | Response Time | Screenshot                                      |
| ---------------------- | ------------------------------------------------------------------------------------- | ------------ | ----------------------------------------------- |
| **Before Using Cache** | Data is still taken directly from the database, so it takes quite a long time. | ⏳ Slow          | ![alt text](</assets/before_using_redis.png>) |
| **After Using Cache**  | Data is taken from Redis Cache so the process becomes faster.                  | ⚡ Fast     | ![alt text](</assets/after_using_redis.png>) |

<br>

🚀 Features
- 🔐 JWT Authentication (Login & Register)
- 🔑 Forgot Password via Email Token
- 🛒 Order Management (Add to Cart, Checkout, Payment)
- 🧾 View Order History & Order Details
- 👤 User Profile Management (Update Personal Information)
- 🛠️ Admin Management for Categories & Products
- ✨ Multiple File Upload (e.g., product images)
- 📘 Swagger Auto-Generated API Documentation
- 🗂️ MVC Architecture
- 📦 PostgreSQL Integration
- 👤 Role-Based Authentication & Authorization (User & Admin)


## 🛠️ Tech Stack
![Go](https://img.shields.io/badge/-Go-00ADD8?logo=go&logoColor=white&style=for-the-badge)
![Gin](https://img.shields.io/badge/-Gin-00ADD8?logo=go&logoColor=white&style=for-the-badge)
![PostgreSQL](https://img.shields.io/badge/-PostgreSQL-4169E1?logo=postgresql&logoColor=white&style=for-the-badge)
![Swagger](https://img.shields.io/badge/Swagger-UI-85EA2D?logo=swagger&logoColor=black&style=for-the-badge)
![JWT](https://img.shields.io/badge/JWT-000000?logo=jsonwebtokens&logoColor=white&style=for-the-badge)
![Argon2](https://img.shields.io/badge/Argon2-0A7E8C?style=for-the-badge)
![Go Migrate](https://img.shields.io/badge/Go%20Migrate-01B3E3?logo=go&logoColor=white&style=for-the-badge)
![Go Mail](https://img.shields.io/badge/Go%20Mail-00ADD8?logo=go&logoColor=white&style=for-the-badge)


##  🔐 .env Configuration
```
# Database
DBUSER=youruser
DBPASS=yourpass
DBHOST=localhost
DBPORT=5432
DBNAME=tickitz

# JWT hash
JWT_SECRET=your_jwt_secret

# Redish
REDISUSER=<redis_user>
REDISPASS=<redis_pass>
REDISPORT=6379
REDISHOST=<redis_host>

# Vercel
DATABASE_URL=<your_url_database>
REDIS_URL=<your_redis_url>

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=<your_email>
SMTP_PASS=<your_app_password_email>
SMTP_FROM=<aplication-name> <your_email> # from
FRONTEND_URL=<your_frontend_url>
```

## 📦 How to Install & Run Project
### 1. First, clone this repository: 
```
https://github.com/federus1105/koda-b4-backend.git
```
### 2. Install Dependencies
```go
go mod tidy
```
### 3. Setup your environment
### 4. Do the Database Migration
### 5. Run Server/Project
```go
go run .\cmd\main.go 
```
### 6. Init Swagger
```go
swag init -g ./cmd/main.go
```
### Open Swagger Documentation in Browser
#### ⚠️ Make sure the server is running
```http://localhost:8011/swagger/index.html```


<br>


## 🗃️ How to run Database Migrations
### ⚠️ Attention: This only applies to PostgreSQL, because enums can only be used in PostgreSQL.
### 1. Install Go migrate
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest;
```
### 2. Create database
```bash
CREATE DATABASE <database_name>;
```
### 3. Migrations Up
```bash
migrate -path ./db/migrations -database "postgres://user:password@localhost:5432/database?sslmode=disable" up
```
### 4. Migrations Down
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

