package models

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserList struct {
	Photo    string `json:"photo"`
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Email    string `json:"email"`
}

type UserBody struct {
	Id        int                   `form:"id"`
	Photos    *multipart.FileHeader `form:"photos" binding:"required"`
	PhotosStr string                `form:"photosStr"`
	Fullname  string                `form:"fullname" binding:"required,max=30"`
	Email     string                `form:"email"  binding:"required,email"`
	Phone     string                `form:"phone" binding:"max=12"`
	Password  string                `form:"password" binding:"required,password_complex"`
	Address   string                `form:"address" binding:"required,max=50"`
	Role      string                `form:"role" binding:"required,oneof=user admin"`
}

func GetListUser(ctx context.Context, db *pgxpool.Pool, name string, limit, offset int) ([]UserList, error) {
	sql := `SELECT 
	COALESCE(photos, '') AS photos, 
	a.fullname, 
	a.phonenumber, 
	COALESCE(address, '') AS address,
	u.email FROM account a
	JOIN users u ON u.id = a.id_users`

	args := []interface{}{}
	argIdx := 1

	// --- SEARCH ---
	if strings.TrimSpace(name) != "" {
		sql += fmt.Sprintf(" WHERE a.fullname ILIKE $%d", argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}

	// --- ORDER LIMIT, OFFSET ---
	sql += fmt.Sprintf(" ORDER BY a.fullname ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var users []UserList
	for rows.Next() {
		var p UserList
		if err := rows.Scan(&p.Photo, &p.Fullname, &p.Phone, &p.Address, &p.Email); err != nil {
			return nil, err
		}
		users = append(users, p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func CreateUser(ctx context.Context, db *pgxpool.Pool, hashPassword string, user UserBody) (UserBody, error) {
	// --- START TRANSACTION ---
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("Failed to begin transaction : ", err)
		return UserBody{}, err
	}
	defer tx.Rollback(ctx)
	var userID int

	// --- INSERT TABLE USER ---
	userSQL := `INSERT INTO users (email, password, role) VALUES ($1, $2, $3) 
		RETURNING id`

	if err := tx.QueryRow(ctx, userSQL,
		user.Email,
		hashPassword,
		user.Role).Scan(&userID); err != nil {
		log.Println("Failed to insert user :", err)
		return UserBody{}, err
	}

	// --- INSERT TABLE ACCOUNT ---
	accountSQL := `INSERT INTO account (id_users, fullname, phoneNumber, address, photos) 
	VALUES ($1, $2, $3, $4, $5) RETURNING id_users, fullname, phoneNumber, address, photos`

	values := []any{userID, user.Fullname, user.Phone, user.Address, user.PhotosStr}
	var newUser UserBody
	if err := tx.QueryRow(ctx, accountSQL, values...).Scan(
		&newUser.Id,
		&newUser.Fullname,
		&newUser.Phone,
		&newUser.Address,
		&newUser.PhotosStr,
	); err != nil {
		log.Println("Failed to insert product:", err)
		return UserBody{}, err
	}

	// --- ASIGN IMAGE STR TO RETURN STRUCT RESPONSE--
	newUser.PhotosStr = user.PhotosStr
	newUser.Email = user.Email
	newUser.Role = user.Role

	// --- COMMIT TRANSACTION ---
	if err = tx.Commit(ctx); err != nil {
		log.Println("Failed to commit Transaction : ", err)
		return UserBody{}, err
	}

	return newUser, nil
}
