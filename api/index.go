package handler

import (
	"fmt"
	"net/http"

	"github.com/federus1105/koda-b4-backend/internals/routes"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var App *gin.Engine
var rd *redis.Client
var db *pgxpool.Pool

func init() {

	App = gin.New()
	App.Use(gin.Recovery())

	routes.InitRouter(App, db, rd)
	App.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"Success": true,
			"Message": "Backend is running 🚀",
		})
	})

	fmt.Println("Router initialized successfully")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Handler called: %s %s\n", r.Method, r.URL.Path)
	App.ServeHTTP(w, r)
}
