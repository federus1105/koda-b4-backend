package handler

import (
	"fmt"
	"net/http"

	"github.com/federus1105/koda-b4-backend/internals/configs"
	"github.com/federus1105/koda-b4-backend/internals/routes"
	"github.com/gin-gonic/gin"
)

var App *gin.Engine

func Handler(w http.ResponseWriter, r *http.Request) {
	if App == nil {
		App = setupApp()
	}
	App.ServeHTTP(w, r)
}

func setupApp() *gin.Engine {

	app := gin.New()
	app.Use(gin.Recovery())

	db, err := configs.ConnectDB()
	if err != nil {
		panic("DB connection failed: " + err.Error())
	}

	rdb, err := configs.NewRedis()
	if err != nil {
		panic("Redis connection failed: " + err.Error())
	}

	routes.InitRouter(app, db, rdb)
	app.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"Success": true,
			"Message": "Backend is running 🚀",
		})
	})

	fmt.Println("Router initialized successfully")
	return app
}
