package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	initDB()

	r := gin.Default()
	r.Use(cors.Default())

	r.Static("/uploads", uploadDirPath())

	api := r.Group("/api")
	{
		api.GET("/ingredients", listIngredients)
		api.POST("/ingredients", createIngredient)
		api.PUT("/ingredients/:id", updateIngredient)
		api.PATCH("/ingredients/:id/stock", setIngredientStock)
		api.DELETE("/ingredients/:id", deleteIngredient)

		api.GET("/dishes", listDishes)
		api.GET("/dishes/:id", getDish)
		api.POST("/dishes", createDish)
		api.PUT("/dishes/:id", updateDish)
		api.PATCH("/dishes/:id/shelf", toggleShelf)
		api.DELETE("/dishes/:id", deleteDish)

		api.GET("/orders", listOrders)
		api.POST("/orders", createOrder)
		api.PATCH("/orders/:id/status", updateOrderStatus)

		api.POST("/upload", uploadImage)
		api.POST("/ai/cooking-method", aiCookingMethod)
		api.POST("/ai/cooking-method-from-url", aiCookingMethodFromURL)
	}

	// Docker / 生产：托管前端构建产物（./static）
	if _, err := os.Stat("./static/index.html"); err == nil {
		serveSPA(r)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("服务启动于 http://0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func serveSPA(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		file := "./static" + path
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			c.File(file)
			return
		}
		c.File("./static/index.html")
	})
}
