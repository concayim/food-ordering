package main

import (
	"time"

	"gorm.io/gorm"
)

// StockInfinite 表示库存无限（不限量）
const StockInfinite = -1

// Ingredient 原材料
type Ingredient struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Unit      string         `gorm:"size:20" json:"unit"` // 单位，如 克/个/份
	Stock     int            `json:"stock"`               // 库存数量，-1 表示无限
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// IsInfinite 是否为无限库存
func (i Ingredient) IsInfinite() bool {
	return i.Stock == StockInfinite
}

// IngredientPurchase 原材料采购流水，用于财务统计
type IngredientPurchase struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	IngredientID uint        `gorm:"index" json:"ingredientId"`
	Ingredient   *Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	Quantity     float64     `json:"quantity"`                          // 本次采购数量
	UnitPrice    float64     `json:"unitPrice"`                         // 单价
	TotalCost    float64     `gorm:"index" json:"totalCost"`            // 本次采购总金额
	PurchaseDate string      `gorm:"size:10;index" json:"purchaseDate"` // YYYY-MM-DD
	Note         string      `gorm:"size:500" json:"note"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

// 菜品类型
const (
	KindDish = "dish" // 菜品
	KindSoup = "soup" // 汤品
)

// Dish 菜品 / 汤品
type Dish struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	Name          string           `gorm:"size:100;not null" json:"name"`
	Description   string           `gorm:"size:500" json:"description"`
	Category      string           `gorm:"size:50" json:"category"`
	Kind          string           `gorm:"size:20;default:'dish'" json:"kind"` // dish 菜品 / soup 汤品
	CookingMethod string           `gorm:"type:text" json:"cookingMethod"`     // 烹饪方法
	ImageURL      string           `gorm:"size:1000" json:"imageUrl"`
	OnShelf       bool             `gorm:"default:false" json:"onShelf"` // 是否上架
	Ingredients   []DishIngredient `gorm:"foreignKey:DishID;constraint:OnDelete:CASCADE" json:"ingredients"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}

// DishIngredient 菜品-原材料关联（每份菜品所需的原材料及用量）
type DishIngredient struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	DishID       uint        `gorm:"index" json:"dishId"`
	IngredientID uint        `json:"ingredientId"`
	Ingredient   *Ingredient `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
	Quantity     float64     `json:"quantity"` // 每份菜品消耗该原材料的数量
}

// Order 订单
type Order struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Status    string      `gorm:"size:20;default:'pending'" json:"status"` // pending/paid/done/cancelled
	Remark    string      `gorm:"size:500" json:"remark"`
	Items     []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

// OrderItem 订单明细
type OrderItem struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	OrderID  uint   `gorm:"index" json:"orderId"`
	DishID   uint   `json:"dishId"`
	DishName string `gorm:"size:100" json:"dishName"`
	Quantity int    `json:"quantity"`
}
