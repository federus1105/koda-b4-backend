package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
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

	// --- CONNECT DATABASE ---
	db, err := configs.ConnectDB()
	if err != nil {
		panic("DB connection failed: " + err.Error())
	}

	// --- CONNECT REDIS ---
	rdb, err := configs.NewRedis()
	if err != nil {
		panic("Redis connection failed: " + err.Error())
	}

	// --- INIT CLOUDINARY ---
	var cld *cloudinary.Cloudinary
	if os.Getenv("CLOUDINARY_URL") != "" {
		cld, err = cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
		if err != nil {
			panic("Cloudinary connection failed: " + err.Error())
		}
		fmt.Println("✅ Cloudinary connected")
	}

	routes.InitRouter(app, db, rdb, cld)
	app.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"Success": true,
			"Message": "Backend is running 🚀",
		})
	})

	fmt.Println("Router initialized successfully")
	return app
}
