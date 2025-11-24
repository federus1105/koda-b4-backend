package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartItemRequest struct {
	ProductID int  `json:"product_id" binding:"gt=0"`
	SizeID    *int `json:"size" binding:"omitempty,gt=0,lte=3"`
	VariantID *int `json:"variant" binding:"omitempty,gt=0,lte=2"`
	Quantity  int  `json:"quantity" binding:"gt=0"`
}

type CartItemResponse struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	ProductID int       `json:"product_id"`
	SizeID    *int      `json:"size_id"`
	VariantID *int      `json:"variant_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Card struct {
	Id            int     `json:"id"`
	Id_product    int     `json:"id_product"`
	Image         string  `json:"images"`
	Name          string  `json:"name"`
	Quantity      int     `json:"qty"`
	Size          *string `json:"size"`
	Variant       *string `json:"variant"`
	Price         float64 `json:"price"`
	PriceDiscount float64 `json:"discount"`
	FlashSale     bool    `json:"flash_sale"`
	Subtotal      float64 `json:"subtotal"`
}

type TransactionsProduct struct {
	Id_product int
	Quantity   int
	Subtotal   float64
	Variant    string
	Size       string
}

type TransactionsInput struct {
	Id_Orders        int     `json:"id_orders"`
	FullName         string  `json:"fullname" binding:"omitempty,max=30"`
	Address          string  `json:"address" binding:"omitempty,max=50"`
	Phone            string  `json:"phone" binding:"omitempty,len=12,numeric"`
	Email            string  `json:"email" binding:"omitempty,email"`
	Id_PaymentMethod int     `json:"id_paymentMethod" binding:"required"`
	Id_Delivery      int     `json:"id_delivery" binding:"required"`
	Order_number     string  `json:"order_number"`
	Subtotal         float64 `json:"subtotal"`
	Tax              int     `json:"tax"`
	DeliveryFee      int     `json:"delivery_fee"`
	Total            float64 `json:"total"`
	Products         []TransactionsProduct
}

func CreateCartProduct(ctx context.Context, db *pgxpool.Pool, accountID int, input CartItemRequest) (*CartItemResponse, error) {
	var stock int
	// --- CHECKING STOCK ---
	err := db.QueryRow(ctx, "SELECT stock FROM product WHERE id = $1", input.ProductID).Scan(&stock)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &CartItemResponse{}, fmt.Errorf("product Not found")
		}
		return nil, fmt.Errorf("failed to check product stock %w", err)
	}

	// --- VALIDATION STOCK ---
	if stock <= 0 {
		return nil, fmt.Errorf("product is out of stock")
	}
	if input.Quantity > stock {
		return nil, fmt.Errorf("insufficient stock %d", stock)
	}

	// --- IF STOCK READY INSERT CART ---
	sql := `INSERT INTO cart (account_id, product_id, size_id, variant_id, quantity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (account_id, product_id, size_id, variant_id)
		DO UPDATE SET 
			quantity = cart.quantity + EXCLUDED.quantity,
			updated_at = NOW()
		RETURNING id, account_id, product_id, size_id, variant_id, quantity, created_at, updated_at;`

	var item CartItemResponse

	err = db.QueryRow(ctx, sql,
		accountID,
		input.ProductID,
		input.SizeID,
		input.VariantID,
		input.Quantity,
	).Scan(
		&item.ID,
		&item.AccountID,
		&item.ProductID,
		&item.SizeID,
		&item.VariantID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return &CartItemResponse{}, fmt.Errorf("failed to insert or update cart item: %w", err)
	}

	return &item, nil
}

func GetCartProduct(ctx context.Context, db *pgxpool.Pool, UserID int) ([]Card, error) {
	sql := `SELECT 
    c.id, 
	p.id as id_product,
    p.name, 
    p.priceoriginal,
    p.pricediscount,
    p.flash_sale,
    pi.photos_one, 
    c.quantity, 
    s.name AS size, 
    v.name AS variant,
    (p.priceoriginal * c.quantity) AS subtotal
FROM cart c
LEFT JOIN sizes s ON s.id = c.size_id
LEFT JOIN variants v ON v.id = c.variant_id
LEFT JOIN product p ON p.id = c.product_id
LEFT JOIN product_images pi ON p.id_product_images = pi.id
WHERE c.account_id = $1;`

	rows, err := db.Query(ctx, sql, UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	carts := []Card{}
	for rows.Next() {
		var c Card
		if err := rows.Scan(
			&c.Id,
			&c.Id_product,
			&c.Name,
			&c.Price,
			&c.PriceDiscount,
			&c.FlashSale,
			&c.Image,
			&c.Quantity,
			&c.Size,
			&c.Variant,
			&c.Subtotal); err != nil {
			return nil, err
		}
		carts = append(carts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return carts, nil
}

func Transactions(ctx context.Context, db *pgxpool.Pool, input TransactionsInput, Iduser int) (TransactionsInput, error) {
	var result TransactionsInput

	// --- GET DATA USER ---
	userData := utils.UserData{
		Email:    input.Email,
		FullName: input.FullName,
		Address:  input.Address,
		Phone:    input.Phone,
	}

	userData, err := utils.GetAndValidateUserData(ctx, db, Iduser, userData)
	if err != nil {
		return result, err
	}

	// Masukkan kembali hasil validasi ke input
	input.Email = userData.Email
	input.FullName = userData.FullName
	input.Address = userData.Address
	input.Phone = userData.Phone

	// --- GET CART USER ---
	rows, err := db.Query(ctx, `
		SELECT c.quantity, p.id as product_id, p.priceoriginal, p.pricediscount, p.flash_sale, s.name AS size, v.name AS variant
		FROM cart c
		JOIN product p ON p.id = c.product_id
		LEFT JOIN sizes s ON s.id = c.size_id
		LEFT JOIN variants v ON v.id = c.variant_id
		WHERE c.account_id=$1
	`, Iduser)
	if err != nil {
		return result, fmt.Errorf("gagal mengambil cart: %v", err)
	}
	defer rows.Close()

	var products []TransactionsProduct
	subtotal := 0.0

	for rows.Next() {
		var productID, quantity int
		var size, variant sql.NullString
		var priceOriginal, priceDiscount float64
		var flashSale bool

		if err := rows.Scan(&quantity, &productID, &priceOriginal, &priceDiscount, &flashSale, &size, &variant); err != nil {
			return result, fmt.Errorf("failed to scan cart items: %v", err)
		}

		price := priceOriginal
		if flashSale {
			price = priceDiscount
		}

		itemSubtotal := price * float64(quantity)
		subtotal += itemSubtotal

		products = append(products, TransactionsProduct{
			Id_product: productID,
			Quantity:   quantity,
			Subtotal:   itemSubtotal,
			Variant:    variant.String,
			Size:       size.String,
		})
	}

	// --- CHECKING CART  ---
	if len(products) == 0 {
		return result, fmt.Errorf("cart is empty, can't place an order")
	}

	// --- GET DELIVERY FEE ---
	var deliveryFee int
	err = db.QueryRow(ctx, `
        SELECT fee FROM delivery WHERE id = $1
    `, input.Id_Delivery).Scan(&deliveryFee)
	if err != nil {
		return result, fmt.Errorf("failed get delivery fee: %v", err)
	}

	// --- TAX ---
	tax := 2000

	// --- TOTAL FINAL ---
	totalFinal := subtotal + float64(deliveryFee) + float64(tax)

	// --- START QUERY TRANSACTION ---
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("Failed to start transaction:", err)
		return TransactionsInput{}, err
	}
	defer tx.Rollback(ctx)

	// --- INSERT ORDERS ---
	var orderID int
	var orderNumber string
	err = tx.QueryRow(ctx, `
		INSERT INTO orders(
			id_account, email, fullname, address, phoneNumber,
			id_delivery, id_paymentmethod, subtotal, tax, delivery_fee,
			total, id_status, createdAt, order_number
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,1,NOW(),
			'#ORD-' || LPAD(nextval('orders_id_seq')::text, 3, '0')
		)
		RETURNING id, order_number
	`,
		Iduser,
		input.Email,
		input.FullName,
		input.Address,
		input.Phone,
		input.Id_Delivery,
		input.Id_PaymentMethod,
		subtotal,
		tax,
		deliveryFee,
		totalFinal,
	).Scan(&orderID, &orderNumber)
	if err != nil {
		return result, fmt.Errorf("failed insert orders: %v", err)
	}

	// --- INSERT PRODUCT ORDERS ---
	for _, p := range products {
		_, err := tx.Exec(ctx, `
			INSERT INTO product_orders(id_order, id_product, quantity, variant, size, subtotal)
			VALUES ($1,$2,$3,$4,$5,$6)
		`, orderID, p.Id_product, p.Quantity, p.Variant, p.Size, p.Subtotal)
		if err != nil {
			return result, fmt.Errorf("failed insert product orders: %v", err)
		}

		// --- UPDATE STOCK ---
		res, err := tx.Exec(ctx, `
        UPDATE product
        SET stock = stock - $1
        WHERE id = $2 AND stock >= $1
    `, p.Quantity, p.Id_product)
		if err != nil {
			return result, fmt.Errorf("failed update stock: %v", err)
		}

		// --- VALIDATION STOCK ---
		if res.RowsAffected() == 0 {
			return result, utils.ValidationError{
				Field:   fmt.Sprintf("product_id_%d", p.Id_product),
				Message: "stock not enough",
			}
		}
	}

	// --- DELETE CART USER ---
	_, err = tx.Exec(ctx, `DELETE FROM cart WHERE account_id=$1`, Iduser)
	if err != nil {
		return result, fmt.Errorf("failed deleted cart: %v", err)
	}

	// --- RESPONSE ---
	result = TransactionsInput{
		Id_Orders:        orderID,
		FullName:         input.FullName,
		Address:          input.Address,
		Phone:            input.Phone,
		Email:            input.Email,
		Id_PaymentMethod: input.Id_PaymentMethod,
		Id_Delivery:      input.Id_Delivery,
		Order_number:     orderNumber,
		Subtotal:         subtotal,
		Tax:              tax,
		DeliveryFee:      deliveryFee,
		Total:            totalFinal,
		Products:         products,
	}

	// --- COMMIT TRANSAKSI ---
	if err := tx.Commit(ctx); err != nil {
		log.Println("failed commmit transaksi:", err)
		return TransactionsInput{}, err
	}

	return result, nil
}

func DeleteCart(ctx context.Context, db *pgxpool.Pool, UserId, IdCart int) error {
	sql := `DELETE FROM cart WHERE account_id = $1 AND id = $2`

	result, err := db.Exec(ctx, sql, UserId, IdCart)
	if err != nil {
		log.Printf("Failed to execute delete, Error: %v", err)
		if ctxErr := ctx.Err(); ctxErr != nil {
			log.Printf("context error: %v", ctxErr)
		}
		return err
	}
	rows := result.RowsAffected()
	log.Printf("Rows affected: %d", rows)

	if rows == 0 {
		return fmt.Errorf("cart with id %d not found", IdCart)
	}

	log.Printf("cart with id %d successfully deleted", IdCart)
	return nil
}
