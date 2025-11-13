package api

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/routes"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	db *pgxpool.Pool
	rd *redis.Client
)

var App *gin.Engine

func init() {
	// === 1. Init PostgreSQL ===
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("⚠️ DATABASE_URL not found in env")
	} else {
		var err error
		db, err = pgxpool.New(context.Background(), dbURL)
		if err != nil {
			fmt.Println("❌ Failed to connect to PostgreSQL:", err)
		} else {
			fmt.Println("✅ Connected to PostgreSQL")
		}
	}

	// === 2. Init Redis ===
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		fmt.Println("⚠️ REDIS_URL not found in env")
	} else {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			fmt.Println("❌ Failed to parse Redis URL:", err)
		} else {
			rd = redis.NewClient(opt)
			fmt.Println("✅ Connected to Redis")
		}
	}

	// === 3. Setup Gin ===
	App = routes.InitRouter(db, rd)

	App.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, models.ResponseSucces{
			Success: true,
			Message: "Backend is running well ✅",
		})
	})

}

func Handler(w http.ResponseWriter, r *http.Request) {
	App.ServeHTTP(w, r)
}
