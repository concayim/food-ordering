package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func getNotifyStatus(c *gin.Context) {
	c.JSON(http.StatusOK, notifierStatus())
}

func testNotify(c *gin.Context) {
	if err := notifyTestMessage(); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "测试消息已发送"})
}

func notifyOrderByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var order Order
	if err := db.Preload("Items").First(&order, id).Error; err != nil {
		fail(c, http.StatusNotFound, "订单不存在")
		return
	}
	st := notifierStatus()
	if !st.Enabled {
		fail(c, http.StatusServiceUnavailable, st.Message)
		return
	}
	notifyOrder(order)
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "已推送到: " + strings.Join(st.Channels, ", ")})
}
