package main

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

// 1. 定義前端傳入的 JSON 結構體 (Struct)
// tag 裡面的 json:"product_id" 代表對應的 JSON 欄位名稱
// binding:"required" 代表這個欄位必填，min=1 代表數量至少要 1 份
type OrderRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// 2. 模擬資料庫庫存 (全域變數)
// 商品 ID 1 (招牌牛腩) 目前庫存剩 5 份
var inventory = map[int]int{
	1: 5, 
}

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "pong"})
	})

	// 3. 建立點餐的 POST 路由
	r.POST("/api/v1/orders", func(c *gin.Context) {
		var req OrderRequest

		// Gin 的強大功能：自動將前端的 JSON 綁定到 req 變數中
		if err := c.ShouldBindJSON(&req); err != nil {
			// 如果前端傳入的格式不對（例如數量小於 1，或欄位缺失）
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "failed",
				"message": "無效的請求資料: " + err.Error(),
			})
			return
		}

		// 4. 核心商業邏輯處理
		currentStock, exists := inventory[req.ProductID]
		// 驚嘆號 ! 代表「非」，!exists 就是「如果不存在」
		// 這裡就能精準抓到：你點了店裡沒賣的東西，直接回傳 404 錯誤！
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "failed",
				"message": "找不到該商品",
			})
			return
		}

		// 檢查庫存是否足夠
		if currentStock < req.Quantity {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":  "failed",
				"message": fmt.Sprintf("很抱歉，庫存不足。目前剩餘: %d 份", currentStock),
			})
			return
		}

		// 扣減庫存 (注意：目前這個寫法在高並發下會有 Bug，這就是我們故意的！)
		inventory[req.ProductID] = currentStock - req.Quantity

		// 回傳成功 JSON
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "點餐成功，已為您預留庫存！",
			"remaining_stock": inventory[req.ProductID],
		})
	})

	r.Run(":8080")
}