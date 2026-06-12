package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"flashorder-go/database"   
	"flashorder-go/repository" 
	"github.com/gin-gonic/gin"
)

type OrderRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

func CreateOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "開啟交易失敗: " + err.Error()})
		return
	}
	defer tx.Rollback()

	// 1. 拿取庫存與版本號
	currentStock, currentVersion, err := repository.GetStockAndVersion(tx, req.ProductID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "找不到該商品"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "查詢出錯"})
		}
		return
	}

	if currentStock < req.Quantity {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status":  "failed",
			"message": fmt.Sprintf("很抱歉，庫存不足。目前剩餘: %d 份", currentStock),
		})
		return
	}

	// 模擬資料庫延遲
	time.Sleep(50 * time.Millisecond)

	// 2. 挑戰寫入資料庫（帶上剛才看到的版本號）
	newStock := currentStock - req.Quantity
	rowsAffected, err := repository.UpdateStockOptimistic(tx, req.ProductID, newStock, currentVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "更新失敗"})
		return
	}

	// 3. 關鍵判斷：如果影響行數是 0，代表版本號已經被別人先改掉了！
	if rowsAffected == 0 {
		fmt.Println("❌【⚡ 樂觀鎖衝突】有人慢了一步，版本號不匹配，寫入被拒絕！")
		c.JSON(http.StatusConflict, gin.H{
			"status":  "failed",
			"message": "系統擁擠，搶單失敗！請重新嘗試。",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "提交失敗"})
		return
	}

	fmt.Printf("【🛡️ 樂觀鎖成功】商品 ID: %d, 成功扣減！殘餘庫存: %d, 新版本號: %d\n", req.ProductID, newStock, currentVersion+1)

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"message":         "點餐成功！",
		"remaining_stock": newStock,
	})
}