package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}

// ---------------- 原材料 Ingredient ----------------

func listIngredients(c *gin.Context) {
	var items []Ingredient
	db.Order("id desc").Find(&items)
	c.JSON(http.StatusOK, items)
}

func createIngredient(c *gin.Context) {
	var in Ingredient
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	in.ID = 0
	if err := db.Create(&in).Error; err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, in)
}

func updateIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ing Ingredient
	if err := db.First(&ing, id).Error; err != nil {
		fail(c, http.StatusNotFound, "原材料不存在")
		return
	}
	var in Ingredient
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	ing.Name = in.Name
	ing.Unit = in.Unit
	ing.Stock = in.Stock
	db.Save(&ing)
	c.JSON(http.StatusOK, ing)
}

// setIngredientStock 设置库存，支持设为无限
func setIngredientStock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		Stock    int  `json:"stock"`
		Infinite bool `json:"infinite"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	var ing Ingredient
	if err := db.First(&ing, id).Error; err != nil {
		fail(c, http.StatusNotFound, "原材料不存在")
		return
	}
	if body.Infinite {
		ing.Stock = StockInfinite
	} else {
		ing.Stock = body.Stock
	}
	db.Save(&ing)
	c.JSON(http.StatusOK, ing)
}

func deleteIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db.Delete(&Ingredient{}, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------------- 菜品 Dish ----------------

func listDishes(c *gin.Context) {
	var dishes []Dish
	q := db.Preload("Ingredients.Ingredient").Order("id desc")
	if c.Query("onShelf") == "true" {
		q = q.Where("on_shelf = ?", true)
	}
	q.Find(&dishes)
	c.JSON(http.StatusOK, dishes)
}

func getDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dish Dish
	if err := db.Preload("Ingredients.Ingredient").First(&dish, id).Error; err != nil {
		fail(c, http.StatusNotFound, "菜品不存在")
		return
	}
	c.JSON(http.StatusOK, dish)
}

type dishInput struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	Kind          string `json:"kind"`
	CookingMethod string `json:"cookingMethod"`
	ImageURL      string `json:"imageUrl"`
	OnShelf       bool   `json:"onShelf"`
	Ingredients   []struct {
		IngredientID uint    `json:"ingredientId"`
		Quantity     float64 `json:"quantity"`
	} `json:"ingredients"`
}

func createDish(c *gin.Context) {
	var in dishInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	kind := in.Kind
	if kind == "" {
		kind = KindDish
	}
	dish := Dish{
		Name: in.Name, Description: in.Description,
		Category: in.Category, Kind: kind, CookingMethod: in.CookingMethod,
		ImageURL: in.ImageURL, OnShelf: in.OnShelf,
	}
	for _, di := range in.Ingredients {
		dish.Ingredients = append(dish.Ingredients, DishIngredient{
			IngredientID: di.IngredientID, Quantity: di.Quantity,
		})
	}
	if err := db.Create(&dish).Error; err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	db.Preload("Ingredients.Ingredient").First(&dish, dish.ID)
	c.JSON(http.StatusOK, dish)
}

func updateDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dish Dish
	if err := db.First(&dish, id).Error; err != nil {
		fail(c, http.StatusNotFound, "菜品不存在")
		return
	}
	var in dishInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	dish.Name = in.Name
	dish.Description = in.Description
	dish.Category = in.Category
	if in.Kind != "" {
		dish.Kind = in.Kind
	}
	dish.CookingMethod = in.CookingMethod
	dish.ImageURL = in.ImageURL
	dish.OnShelf = in.OnShelf

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&dish).Error; err != nil {
			return err
		}
		// 重建原材料关联
		if err := tx.Where("dish_id = ?", dish.ID).Delete(&DishIngredient{}).Error; err != nil {
			return err
		}
		for _, di := range in.Ingredients {
			if err := tx.Create(&DishIngredient{
				DishID: dish.ID, IngredientID: di.IngredientID, Quantity: di.Quantity,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	db.Preload("Ingredients.Ingredient").First(&dish, dish.ID)
	c.JSON(http.StatusOK, dish)
}

// toggleShelf 上架/下架
func toggleShelf(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dish Dish
	if err := db.First(&dish, id).Error; err != nil {
		fail(c, http.StatusNotFound, "菜品不存在")
		return
	}
	dish.OnShelf = !dish.OnShelf
	db.Save(&dish)
	c.JSON(http.StatusOK, dish)
}

func deleteDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db.Select("Ingredients").Delete(&Dish{ID: uint(id)})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------------- 订单 Order ----------------

type orderInput struct {
	Remark string `json:"remark"`
	Items  []struct {
		DishID   uint `json:"dishId"`
		Quantity int  `json:"quantity"`
	} `json:"items"`
}

func createOrder(c *gin.Context) {
	var in orderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if len(in.Items) == 0 {
		fail(c, http.StatusBadRequest, "订单不能为空")
		return
	}

	var created Order
	err := db.Transaction(func(tx *gorm.DB) error {
		order := Order{Status: "pending", Remark: in.Remark}

		// 累计每种原材料的需求量
		needed := map[uint]float64{}

		for _, item := range in.Items {
			if item.Quantity <= 0 {
				continue
			}
			var dish Dish
			if err := tx.Preload("Ingredients").First(&dish, item.DishID).Error; err != nil {
				return errors.New("菜品不存在")
			}
			if !dish.OnShelf {
				return errors.New("菜品「" + dish.Name + "」已下架")
			}
			for _, di := range dish.Ingredients {
				needed[di.IngredientID] += di.Quantity * float64(item.Quantity)
			}
			order.Items = append(order.Items, OrderItem{
				DishID: dish.ID, DishName: dish.Name, Quantity: item.Quantity,
			})
		}

		// 校验并扣减库存（无限库存跳过）
		for ingID, qty := range needed {
			var ing Ingredient
			if err := tx.First(&ing, ingID).Error; err != nil {
				return errors.New("原材料不存在")
			}
			if ing.IsInfinite() {
				continue
			}
			if float64(ing.Stock) < qty {
				return errors.New("原材料「" + ing.Name + "」库存不足")
			}
			ing.Stock -= int(qty)
			if err := tx.Save(&ing).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		created = order
		return nil
	})
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	notifyOrderAsync(created.ID)
	c.JSON(http.StatusOK, created)
}

func listOrders(c *gin.Context) {
	var orders []Order
	db.Preload("Items").Order("id desc").Find(&orders)
	c.JSON(http.StatusOK, orders)
}

func updateOrderStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	var order Order
	if err := db.First(&order, id).Error; err != nil {
		fail(c, http.StatusNotFound, "订单不存在")
		return
	}
	order.Status = body.Status
	db.Save(&order)
	c.JSON(http.StatusOK, order)
}
