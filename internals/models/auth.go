package models

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	DB  *pgxpool.Pool
	RDB *redis.Client
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

type Users struct {
	Id       int    `json:"id"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"password_complex"`
}

type ReqResetPassword struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"password_complex"`
}

type ReqForgot struct {
	Email string `json:"email" binding:"required,email"`
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

// ----- GET USER BY EMAIL -----
func GetUserByEmail(ctx context.Context, db *pgxpool.Pool, email string) (*Users, error) {
	query := "SELECT id, email, password FROM users WHERE email=$1"
	var user Users
	if err := db.QueryRow(ctx, query, email).Scan(&user.Id, &user.Email, &user.Password); err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// --- SAVE TOKEN FOR RESET PASSWORD ---
func SaveResetToken(ctx context.Context, rdb *redis.Client, key, userID string, ttl time.Duration) error {
	return rdb.Set(ctx, key, userID, ttl).Err()
}

// --- GET USER ID FROM TOKEN REDIS ---
func GetUserIDFromToken(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", errors.New("invalid or expired token")
	}
	return val, nil
}

// --- UPDATE PASSWORD ---
func UpdatePasswordByID(ctx context.Context, db *pgxpool.Pool, userID int, hashedPassword string) error {
	_, err := db.Exec(ctx,
		"UPDATE users SET password = $1 WHERE id = $2",
		hashedPassword, userID,
	)
	return err
}
