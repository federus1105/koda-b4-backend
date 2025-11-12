package utils

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserData struct {
	Email    string
	FullName string
	Address  string
	Phone    string
}

type ValidationError struct {
	Field   string
	Message string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s %s", v.Field, v.Message)
}

func GetAndValidateUserData(ctx context.Context, db *pgxpool.Pool, userID int, input UserData) (UserData, error) {
	var (
		userEmail, userFullName, userAddress, userPhone sql.NullString
	)
	// --- GET USER ID --
	err := db.QueryRow(ctx, `
			SELECT u.email, a.fullname, a.address, a.phonenumber 
			FROM account a
			JOIN users u ON u.id = a.id_users
			WHERE a.id = $1;
		`, userID).Scan(&userEmail, &userFullName, &userAddress, &userPhone)

	if err != nil {
		return input, fmt.Errorf("failed to get user data: %v", err)
	}

	// --- HELPER FUNCTION VALIDATION DATA ---
	getValue := func(inputVal string, userVal sql.NullString, fieldName string) (string, error) {
		if inputVal != "" {
			return inputVal, nil
		}
		if userVal.Valid && userVal.String != "" {
			return userVal.String, nil
		}
		return "", ValidationError{
			Field:   fieldName,
			Message: "field is required",
		}
	}

	// -- VALIDATION & INPUT DB FIELD NULL ---
	if input.Email, err = getValue(input.Email, userEmail, "email"); err != nil {
		return input, err
	}
	if input.FullName, err = getValue(input.FullName, userFullName, "fullname"); err != nil {
		return input, err
	}
	if input.Address, err = getValue(input.Address, userAddress, "address"); err != nil {
		return input, err
	}
	if input.Phone, err = getValue(input.Phone, userPhone, "phone"); err != nil {
		return input, err
	}

	return input, nil
}
