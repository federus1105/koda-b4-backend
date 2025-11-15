package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	docs "github.com/federus1105/koda-b4-backend/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(db *pgxpool.Pool, rd *redis.Client) *gin.Engine {
	router := gin.Default()
	utils.InitValidator()
	router.Use(gin.Recovery())

	// --- SWAGGER ---
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Static("/img", "public")

	// --- ROUTE ---
	InitAuthRouter(router, db)
	InitProductRouter(router, db, rd)
	InitOrderRouter(router, db)
	InitUserRoute(router, db)
	InitCategoriesRouter(router, db)
	InitOrderClientRoutes(router, db)
	InitHistoryRouter(router, db)
	InitProfileRouter(router, db)

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Route Not Found, Try Again!",
		})
	})
	return router

}
