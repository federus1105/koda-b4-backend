package handler

import (
	"fmt"
	"net/http"

	"github.com/federus1105/koda-b4-backend/internals/configs"
	"github.com/federus1105/koda-b4-backend/internals/routes"
	"github.com/gin-gonic/gin"
)

var App *gin.Engine

func init() {

	App = gin.New()
	App.Use(gin.Recovery())

	db, err := configs.ConnectDB()
	if err != nil {
		panic("DB connection failed: " + err.Error())
	}

	rdb := configs.NewRedis()

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
	fmt.Printf("Handler called: %s %s\n", r.Method, r.URL.Path)
	App.ServeHTTP(w, r)
}
