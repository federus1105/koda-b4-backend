package models

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileUpdate struct {
	Id        int                   `form:"id"`
	Fullname  *string               `form:"fullname" binding:"omitempty,max=30"`
	Email     *string               `form:"email" binding:"omitempty,email"`
	Phone     *string           	`form:"phone" binding:"omitempty,min=10,max=13,numeric"`
	Address   *string               `form:"address" binding:"omitempty,max=50"`
	Photos    *multipart.FileHeader `form:"photos"`
	PhotosStr *string               `form:"photosStr,omitempty"`
}

type ReqUpdatePassword struct {
	OldPassword     string `json:"old_password"  binding:"required"`
	NewPassword     string `json:"new_password"  binding:"required,password_complex"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type Profiles struct {
	Id        int       `json:"id"`
	Fullname  string    `json:"fullname"`
	Phone     *string   `json:"phone"`
	Address   *string   `json:"address"`
	Photos    *string   `json:"photos"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
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

	if input.Phone != nil && *input.Phone != "" {
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

func UpdatePassword(ctx context.Context, db *pgxpool.Pool, userID int, oldPassword, newPassword string) error {
	// --- GET PASSWORD ---
	var hashedDB string
	err := db.QueryRow(ctx, "SELECT password FROM users WHERE id = $1", userID).Scan(&hashedDB)
	if err != nil {
		log.Println("Failed to get current password hash:", err)
		return fmt.Errorf(" %w user not found", utils.ErrValidation)
	}

	// --- VERIFICATION OLD PASSWORD ---
	ok, err := libs.VerifyPassword(oldPassword, hashedDB)
	if err != nil {
		log.Println("Error verifying old password:", err)
		return fmt.Errorf("%w, failed verification old password", utils.ErrValidation)
	}
	if !ok {
		return fmt.Errorf("%w the old password doesn't match", utils.ErrValidation)
	}

	// --- HASH NEW PASSWORD ---
	newHashed, err := libs.HashPassword(newPassword)
	if err != nil {
		log.Println("Failed to hash new password:", err)
		return fmt.Errorf("%w failed hash new password", utils.ErrValidation)
	}

	// --- UPDATE PASSWORD ---
	cmdTag, err := db.Exec(ctx, "UPDATE users SET password = $1 WHERE id = $2", newHashed, userID)
	if err != nil {
		log.Println("Failed to update password:", err)
		return fmt.Errorf("failed update password")
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf(" %w failed update password, user not found", utils.ErrValidation)
	}
	return nil
}

func Profile(ctx context.Context, db *pgxpool.Pool, userId int) (Profiles, error) {
	var profile Profiles
	sql := `SELECT a.id, a.fullname, 
	a.phonenumber, a.address, 
	a.photos, u.email,
	a.createdat FROM account a
	JOIN users u ON u.id = a.id_users
	WHERE u.id = $1`

	err := db.QueryRow(ctx, sql, userId).Scan(
		&profile.Id,
		&profile.Fullname,
		&profile.Phone,
		&profile.Address,
		&profile.Photos,
		&profile.Email,
		&profile.CreatedAt,
	)

	if err != nil {
		return profile, err
	}

	return profile, nil
}
