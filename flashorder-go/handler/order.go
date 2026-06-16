package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"flashorder-go/database"   // 💡 如果你的專案資料夾名稱不同，記得修改這裡（例如 sideproject/database）
	"flashorder-go/repository" // 💡 記得根據你的專案名稱修改路徑
	"github.com/gin-gonic/gin"
)

type OrderRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// =========================================================================
// 1. 核心秒殺下單 API
// =========================================================================
func CreateOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": err.Error()})
		return
	}

	ctx := context.Background()
	redisKey := fmt.Sprintf("product:%d:stock", req.ProductID)

	// 【第一道防線】直接在 Redis 記憶體裡把庫存減 1
	redisStock, err := database.RDB.Decr(ctx, redisKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "快取系統異常: " + err.Error()})
		return
	}

	// 如果扣減後的數值小於 0，代表沒庫存了
	if redisStock < 0 {
		database.RDB.Incr(ctx, redisKey) // 補償機制
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status":  "failed",
			"message": "很抱歉，招牌牛腩已被搶購一空！(由 Redis 攔截)",
		})
		return
	}

	fmt.Printf("🎉【Redis 搶單成功】恭喜進入決賽圈！當前記憶體殘餘庫存: %d\n", redisStock)

	// 【第二道後方基地】幫這 5 個幸運兒去 SQLite 記帳
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "開啟交易失敗"})
		return
	}
	defer tx.Rollback()

	// 讀取當前 SQLite 實體庫存
	currentStock, err := repository.GetStockForUpdate(tx, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "讀取總帳失敗"})
		return
	}

	// 模擬實體寫入硬碟的延遲
	time.Sleep(50 * time.Millisecond)

	// 直接更新 SQLite 總帳
	newStock := currentStock - req.Quantity
	err = repository.UpdateStock(tx, req.ProductID, newStock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "寫入總帳失敗"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "提交總帳失敗"})
		return
	}

	fmt.Printf("【💾 總帳寫入完畢】商品 ID: %d, SQLite 庫存同步更新為: %d\n", req.ProductID, newStock)

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"message":         "秒殺成功！已為您保留餐點。",
		"remaining_stock": redisStock,
	})
} // 👈 剛剛就是漏掉了這個用來結束 CreateOrder 的大括號！

// =========================================================================
// 2. 供前端網頁即時查詢 Redis 庫存的 API
// =========================================================================
func GetStock(c *gin.Context) {
	ctx := context.Background()
	redisStock, err := database.RDB.Get(ctx, "product:1:stock").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "讀取快取失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"stock":  redisStock,
	})
}