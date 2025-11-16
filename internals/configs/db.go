package configs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB() (*pgxpool.Pool, error) {
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASS")
	dbHost := os.Getenv("DBHOST")
	dbPort := os.Getenv("DBPORT")
	dbName := os.Getenv("DBNAME")
	connstring := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	log.Println(connstring)
	return pgxpool.New(context.Background(), connstring)
}

func ConnectDB() (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	return pgxpool.New(context.Background(), url)
}
