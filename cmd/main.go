package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/federus1105/koda-b4-backend/internals/configs"
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Coffeeshop Senja Kopi Kiri
// @version 1.0
// @description This is a Documentation API for Coffeshop Senja Kopi Kiri
// @host localhost:8011
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
func main() {
	router := gin.Default()
	router.Use(gin.Recovery())

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
	log.Println("✅ DB Connected boy")

	// --- INIT RDB ---
	rdb, Rdb, err := configs.InitRedis()
	if err != nil {
		log.Println("❌ Failed to connect to redis\nCause: ", err.Error())
		return
	}
	defer rdb.Close()
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		fmt.Println("Failed Connected Redis : ", err.Error())
		return
	}
	log.Println("✅ REDIS Connected boy: ", Rdb)

	// --- INIT CLAUDINARY ---
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		log.Println("Failed Connected cloudinary : ", err)
	}

	if cld != nil {
		log.Println("✅ Cloudinary connected")
	} else {
		log.Println("⚠ Cloudinary not initialized")
	}

	router.Group("/")

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, models.ResponseSucces{
			Success: true,
			Message: "Backend is running boy",
		})
	})

	routes.InitRouter(router, db, rdb, cld)
	router.Run(":8011")
}
