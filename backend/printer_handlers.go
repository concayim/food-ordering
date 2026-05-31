package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// getPrinterStatus 查询打印机驱动状态
func getPrinterStatus(c *gin.Context) {
	c.JSON(http.StatusOK, printer.Status())
}

// getOrderReceipt 预览订单小票内容（不触发打印）
func getOrderReceipt(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	receipt, err := loadReceipt(uint(id))
	if err != nil {
		fail(c, http.StatusNotFound, "订单不存在")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"receipt": receipt,
		"content": formatReceiptText(receipt),
	})
}

// printOrder 提交订单到打印机（通过已注册的 PrinterDriver）
func printOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	receipt, err := loadReceipt(uint(id))
	if err != nil {
		fail(c, http.StatusNotFound, "订单不存在")
		return
	}
	st := printer.Status()
	if !st.Ready {
		fail(c, http.StatusServiceUnavailable, st.Message)
		return
	}
	result := printer.Print(receipt)
	if !result.Success {
		fail(c, http.StatusInternalServerError, result.Message)
		return
	}
	c.JSON(http.StatusOK, result)
}

// testPrint 测试打印（用于驱动联调）
func testPrint(c *gin.Context) {
	st := printer.Status()
	if !st.Ready {
		fail(c, http.StatusServiceUnavailable, st.Message)
		return
	}
	receipt := Receipt{
		OrderID:   0,
		ShopName:  shopName(),
		Status:    "test",
		Remark:    "打印机联调测试",
		Items:     []ReceiptLine{{Name: "测试菜品", Quantity: 1}},
		CreatedAt: time.Now(),
	}
	result := printer.Print(receipt)
	c.JSON(http.StatusOK, result)
}
