package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitHistoryRouter(router *gin.Engine, db *pgxpool.Pool) {
	historyRouter := router.Group("/history")

	historyRouter.GET("", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.GetHistory(ctx, db)
	})

	historyRouter.GET("/:id", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.DetailHistory(ctx, db)
	})
}
