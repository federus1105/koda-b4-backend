package routes

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	docs "github.com/federus1105/koda-b4-backend/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(app *gin.Engine, db *pgxpool.Pool, rd *redis.Client, cld *cloudinary.Cloudinary) {
	utils.InitValidator()

	// --- SWAGGER ---
	docs.SwaggerInfo.BasePath = "/"
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	app.Static("/img", "public")

	// --- ROUTE ---
	InitAuthRouter(app, db, rd)
	InitProductRouter(app, db, rd, cld)
	InitOrderRouter(app, db)
	InitUserRoute(app, db)
	InitCategoriesRouter(app, db)
	InitOrderClientRoutes(app, db)
	InitHistoryRouter(app, db)
	InitProfileRouter(app, db)

	app.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Route Not Found, Try Again!",
		})
	})

}
