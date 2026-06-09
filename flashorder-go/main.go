package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化 Gin 預設引擎
	r := gin.Default()

	// 2. 建立一個簡單的測試網址 (GET /ping)
	r.GET("/ping", func(c *gin.Context) {
		// 回傳狀態碼 200 與 JSON 資料
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "pong! 伺服器運作正常！",
		})
	})

	// 3. 讓伺服器監聽並運行在本機的 8080 連接埠 (Port)
	// 啟動後可以透過 http://localhost:8080/ping 造訪
	r.Run(":8080")
}