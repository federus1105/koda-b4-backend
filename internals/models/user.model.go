package models

import (
	"context"
	"fmt"
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
