package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	initDB()

	r := gin.Default()
	r.Use(cors.Default())

	// 上传图片的静态访问
	r.Static("/uploads", "./"+uploadDir)

	api := r.Group("/api")
	{
		// 原材料
		api.GET("/ingredients", listIngredients)
		api.POST("/ingredients", createIngredient)
		api.PUT("/ingredients/:id", updateIngredient)
		api.PATCH("/ingredients/:id/stock", setIngredientStock)
		api.DELETE("/ingredients/:id", deleteIngredient)

		// 菜品
		api.GET("/dishes", listDishes)
		api.GET("/dishes/:id", getDish)
		api.POST("/dishes", createDish)
		api.PUT("/dishes/:id", updateDish)
		api.PATCH("/dishes/:id/shelf", toggleShelf)
		api.DELETE("/dishes/:id", deleteDish)

		// 订单
		api.GET("/orders", listOrders)
		api.POST("/orders", createOrder)
		api.PATCH("/orders/:id/status", updateOrderStatus)

		// 图片上传 & AI
		api.POST("/upload", uploadImage)
		api.POST("/ai/cooking-method", aiCookingMethod)
		api.POST("/ai/cooking-method-from-url", aiCookingMethodFromURL)
	}

	log.Println("后端启动于 http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
