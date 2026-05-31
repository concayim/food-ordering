package main

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	if err := os.MkdirAll(dataDir(), 0o755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}
	db, err = gorm.Open(sqlite.Open(dbPath()), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := db.AutoMigrate(&Ingredient{}, &Dish{}, &DishIngredient{}, &Order{}, &OrderItem{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	seedData()
}

// seedData 首次运行时写入演示数据
func seedData() {
	var count int64
	db.Model(&Ingredient{}).Count(&count)
	if count > 0 {
		return
	}

	ingredients := []Ingredient{
		{Name: "西红柿", Unit: "个", Stock: 50},
		{Name: "鸡蛋", Unit: "个", Stock: 100},
		{Name: "牛肉", Unit: "克", Stock: 5000},
		{Name: "面条", Unit: "份", Stock: 80},
		{Name: "紫菜", Unit: "克", Stock: 2000},
		{Name: "食用油", Unit: "毫升", Stock: StockInfinite}, // 无限库存
		{Name: "食盐", Unit: "克", Stock: StockInfinite},
		{Name: "米饭", Unit: "份", Stock: StockInfinite},
		{Name: "清水", Unit: "毫升", Stock: StockInfinite},
	}
	db.Create(&ingredients)

	idOf := func(name string) uint {
		for _, ing := range ingredients {
			if ing.Name == name {
				return ing.ID
			}
		}
		return 0
	}

	dishes := []Dish{
		{
			Name: "西红柿炒鸡蛋", Description: "经典家常菜，酸甜可口", Category: "热菜", Kind: KindDish,
			CookingMethod: "1. 西红柿切块，鸡蛋打散加少许盐；\n2. 热油先炒鸡蛋盛出；\n3. 下西红柿炒出汁，倒入鸡蛋翻炒，调味即可。",
			ImageURL:      "https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=400", OnShelf: true,
			Ingredients: []DishIngredient{
				{IngredientID: idOf("西红柿"), Quantity: 2},
				{IngredientID: idOf("鸡蛋"), Quantity: 3},
				{IngredientID: idOf("食用油"), Quantity: 20},
				{IngredientID: idOf("食盐"), Quantity: 5},
			},
		},
		{
			Name: "红烧牛肉面", Description: "汤浓肉烂，分量十足", Category: "主食", Kind: KindDish,
			CookingMethod: "1. 牛肉焯水后炖煮至软烂；\n2. 面条煮熟过水；\n3. 牛肉汤浇面，加葱花调味。",
			ImageURL:      "https://images.unsplash.com/photo-1569718212165-3a8278d5f624?w=400", OnShelf: true,
			Ingredients: []DishIngredient{
				{IngredientID: idOf("牛肉"), Quantity: 200},
				{IngredientID: idOf("面条"), Quantity: 1},
				{IngredientID: idOf("食盐"), Quantity: 8},
			},
		},
		{
			Name: "米饭", Description: "东北珍珠米", Category: "主食", Kind: KindDish,
			ImageURL: "https://images.unsplash.com/photo-1516684732162-798a0062be99?w=400", OnShelf: true,
			Ingredients: []DishIngredient{
				{IngredientID: idOf("米饭"), Quantity: 1},
			},
		},
		{
			Name: "紫菜蛋花汤", Description: "清淡鲜美，暖胃首选", Category: "汤品", Kind: KindSoup,
			CookingMethod: "1. 清水烧开，放入紫菜；\n2. 鸡蛋打散沿锅边淋入成蛋花；\n3. 加盐和几滴香油调味出锅。",
			ImageURL:      "https://images.unsplash.com/photo-1547592180-85f173990554?w=400", OnShelf: true,
			Ingredients: []DishIngredient{
				{IngredientID: idOf("紫菜"), Quantity: 10},
				{IngredientID: idOf("鸡蛋"), Quantity: 1},
				{IngredientID: idOf("清水"), Quantity: 500},
				{IngredientID: idOf("食盐"), Quantity: 3},
			},
		},
	}
	db.Create(&dishes)
	log.Println("已写入演示数据")
}
