package handler

import (
	"net/http"

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
	App = gin.New()
	App.Use(gin.Recovery())

	router := App.Group("/")

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, models.ResponseSucces{
			Success: true,
			Message: "Backend is running well",
		})
	})
	routes.InitRouter(db, rd)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	App.ServeHTTP(w, r)
}
