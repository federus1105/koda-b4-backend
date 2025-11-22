package routes

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitProfileRouter(router *gin.Engine, db *pgxpool.Pool, cld *cloudinary.Cloudinary) {
	profileRouter := router.Group("/profile")

	profileRouter.PATCH("", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.ProfileUpdate(ctx, db, cld)
	})

	profileRouter.PUT("", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.UpdatePassword(ctx, db)
	})

	profileRouter.GET("", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.Profile(ctx, db)
	})
}
