package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/federus1105/koda-b4-backend/internals/configs"
	"github.com/federus1105/koda-b4-backend/internals/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var App *gin.Engine

func init() {
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}

	gin.SetMode(gin.ReleaseMode) 
	App = gin.New()
	App.Use(gin.Recovery())

	db, err := configs.ConnectDB()
	if err != nil {
		panic("DB connection failed: " + err.Error())
	}

	rdb, _, err := configs.InitRedis()
	if err != nil {
		panic("Redis connection failed: " + err.Error())
	}

	routes.InitRouter(App, db, rdb)
	App.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"Success": true,
			"Message": "Backend is running 🚀",
		})
	})

	fmt.Println("Router initialized successfully")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	App.ServeHTTP(w, r)
}
