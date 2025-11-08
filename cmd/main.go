package main

import (
	"log"
	"os"

	"github.com/federus1105/koda-b4-backend/internals/configs"
	"github.com/federus1105/koda-b4-backend/internals/routes"
	"github.com/joho/godotenv"
)

func main() {
	// --- LOAD .ENV IF DEVELOPMENT ---
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}

	// --- INIT DB ---
	db, err := configs.InitDB()
	if err != nil {
		log.Println("❌ Failed to connect to database\nCause: ", err.Error())
		return
	}
	
	defer db.Close()
	log.Println("✅ DB Connected")

	router := routes.InitRouter(db)
	router.Run("localhost:8011")
}
