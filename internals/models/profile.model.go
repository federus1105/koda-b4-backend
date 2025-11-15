package models

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileUpdate struct {
	Id        int                   `form:"id"`
	Fullname  *string               `form:"fullname" binding:"omitempty,max=30"`
	Email     *string               `form:"email" binding:"omitempty,email"`
	Phone     *string               `form:"phone" binding:"omitempty,len=12,numeric"`
	Address   *string               `form:"address" binding:"omitempty,max=50"`
	Photos    *multipart.FileHeader `form:"photos"`
	PhotosStr *string               `form:"photosStr,omitempty"`
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

	// --- UPDATE EMAIL DI USERS TABLE ---
	if input.Email != nil {
		queryEmail := `UPDATE users SET email=$1 WHERE id=$2`
		_, err := tx.Exec(ctx, queryEmail, *input.Email, Id)
		if err != nil {
			log.Println("Failed to update email:", err)
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
