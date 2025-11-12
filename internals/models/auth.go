package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRegister struct {
	Id       int    `json:"id"`
	Fullname string `json:"fullname" binding:"required,max=20"`
	Email    string `json:"email"  binding:"required,email"`
	Password string `json:"password" binding:"required,password_complex"`
}

type AuthLogin struct {
	Id       int    `json:"id"`
	Email    string `json:"email"  binding:"required,email"`
	Password string `json:"password" binding:"required,password_complex"`
	Role     string `json:"role,omitempty"`
}

type ProfileUpdate struct {
	Id        int                   `form:"id"`
	Fullname  *string               `form:"fullname" binding:"omitempty,max=30"`
	Email     *string               `form:"email" binding:"omitempty,email"`
	Phone     *string               `form:"phone" binding:"omitempty,max=12"`
	Address   *string               `form:"address" binding:"omitempty,max=50"`
	Photos    *multipart.FileHeader `form:"photos"`
	PhotosStr *string               `form:"photosStr,omitempty"`
}

func Register(ctx context.Context, db *pgxpool.Pool, hashed string, user AuthRegister) (AuthRegister, error) {
	// --- START TRANSACTION ---
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("Failed to begin transaction : ", err)
		return AuthRegister{}, err
	}
	defer tx.Rollback(ctx)
	var userID int

	// --- INSERT TABLE USER ---
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password) VALUES ($1, $2) 
		RETURNING id`, user.Email, hashed).Scan(&userID)
	if err != nil {
		log.Println("Failed to insert user :", err)
		return AuthRegister{}, err
	}

	// --- INSERT TABLE ACCOUNT ---
	_, err = tx.Exec(ctx,
		`INSERT INTO account (id_users, fullname) VALUES ($1, $2)`,
		userID, user.Fullname)
	if err != nil {
		log.Println("Failed to insert account : ", err)
		return AuthRegister{}, err
	}

	// --- COMMIT TRANSACTION ---
	if err = tx.Commit(ctx); err != nil {
		log.Println("Failed to commit Transaction : ", err)
		return AuthRegister{}, err
	}

	return AuthRegister{
		Id:       userID,
		Email:    user.Email,
		Fullname: user.Fullname,
	}, nil
}

func Login(ctx context.Context, db *pgxpool.Pool, email string) (AuthLogin, error) {
	sql := `SELECT id, email, password, role FROM users WHERE email = $1`
	var user AuthLogin
	if err := db.QueryRow(ctx, sql, email).Scan(&user.Id, &user.Email, &user.Password, &user.Role); err != nil {
		if err == pgx.ErrNoRows {
			return AuthLogin{}, errors.New("user not found")
		}
		log.Println("Internal Server Error.\nCause: ", err.Error())
		return AuthLogin{}, err
	}
	return user, nil
}

func UpdateProfile(ctx context.Context, db *pgxpool.Pool, input ProfileUpdate, Id int) (ProfileUpdate, error) {
	// --- START QUERY TRANSACTION ---
	tx, err := db.Begin(ctx)
	if err != nil {
		return ProfileUpdate{}, err
	}
	defer tx.Rollback(ctx)

	// --- DYNAMIC SET USER ---
	setClauses := []string{}
	args := []any{}
	idx := 1

	if input.Fullname != nil {
		setClauses = append(setClauses, fmt.Sprintf("fullname=$%d", idx))
		args = append(args, *input.Fullname)
		idx++
	}

	if input.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phoneNumber=$%d", idx))
		args = append(args, *input.Phone)
		idx++
	}

	if input.Address != nil {
		setClauses = append(setClauses, fmt.Sprintf("address=$%d", idx))
		args = append(args, *input.Address)
		idx++
	}

	if input.PhotosStr != nil {
		setClauses = append(setClauses, fmt.Sprintf("photos=$%d", idx))
		args = append(args, *input.PhotosStr)
		idx++
	}

	if len(setClauses) > 0 {
		query := fmt.Sprintf("UPDATE account SET %s WHERE id=$%d", strings.Join(setClauses, ","), idx)
		args = append(args, Id)
		_, err := tx.Exec(ctx, query, args...)
		if err != nil {
			log.Println("Failed to update account:", err)
			return ProfileUpdate{}, err
		}
	}

	// --- GET UPDATED DATA ---
	var updated ProfileUpdate
	querySelect := `
		SELECT 
		u.id,
        u.email,
		COALESCE(a.fullname, '-'),
		COALESCE(a.phoneNumber, '-'),
		COALESCE(a.address, '-'),
		COALESCE(a.photos, '-')
	FROM account a
	JOIN users u ON u.id = a.id_users
	WHERE a.id = $1
	`
	if err := tx.QueryRow(ctx, querySelect, Id).Scan(
		&updated.Id,
		&updated.Email,
		&updated.Fullname,
		&updated.Phone,
		&updated.Address,
		&updated.PhotosStr,
	); err != nil {
		log.Println("Failed to fetch updated user:", err)
		return ProfileUpdate{}, err
	}

	// --- COMMIT TRANSACTION ---
	if err := tx.Commit(ctx); err != nil {
		log.Println("Failed to commit transaction:", err)
		return ProfileUpdate{}, err
	}

	return updated, nil
}
