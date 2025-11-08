package models

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRegister struct {
	Id       int    `json:"id"`
	Fullname string `json:"fullname" binding:"required,max=20"`
	Email    string `json:"email"  binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(ctx context.Context, db *pgxpool.Pool, email, hashedPassword, fullname string) (AuthRegister, error) {
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
		RETURNING id`, email, hashedPassword).Scan(&userID)
	if err != nil {
		log.Println("Failed to insert user :", err)
		return AuthRegister{}, err
	}

	// --- INSERT TABLE ACCOUNT ---
	_, err = tx.Exec(ctx,
		`INSERT INTO account (id_users, fullname) VALUES ($1, $2)`,
		userID, fullname)
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
		Email:    email,
		Fullname: fullname,
	}, nil
}
