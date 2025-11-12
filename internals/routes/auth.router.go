package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitAuthRouter(router *gin.Engine, db *pgxpool.Pool) {
	authRouter := router.Group("/auth")
	profileRouter := router.Group("")

	authRouter.POST("/register", func(ctx *gin.Context) {
		controllers.Register(ctx, db)
	})
	authRouter.POST("/login", func(ctx *gin.Context) {
		controllers.Login(ctx, db)
	})

	profileRouter.PATCH("/profile", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.ProfileUpdate(ctx, db)
	})

}
