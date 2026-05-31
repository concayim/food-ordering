package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// uploadImage 接收 multipart 文件（拍照/选图），保存到本地并返回可访问的 URL
func uploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "未收到文件: "+err.Error())
		return
	}

	dir := uploadDirPath()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".png"
	}
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(dir, name)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + name})
}
