package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitAuthRouter(router *gin.Engine, db *pgxpool.Pool) {
	authRouter := router.Group("/auth")
	authRouter.POST("/register", controllers.Register)
}
