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

	// 🌟【關鍵新增】從警衛貼在請求胸口的備忘錄裡，撈出 user_id
	// c.Get 回傳的是空介面，我們要用 .(int) 斷言它是一個整數型態（這就是 Go 的強型別規範）
	rawUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "無法識別會員身分"})
		return
	}
	userID := rawUserID.(int) // 正式轉職為 Go 的 int 變數

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

	// 1. 更新實體總帳庫存
	newStock := currentStock - req.Quantity
	err = repository.UpdateStock(tx, req.ProductID, newStock)
	if err != nil {
		// 🚨【補償機制】Postgres 扣庫存失敗了，立刻把 Redis 的庫存加回來！
		// database.RedisClient.IncrBy 負責把指定的 key 加上指定的數量
		database.RedisClient.IncrBy(c.Request.Context(),"product:1:stock",int64(req.Quantity))
		tx.Rollback() // 資料庫回滾
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "寫入總帳失敗，已回滾快取"})
		return	
	}
	
	// 🌟【關鍵新增】2. 建立訂單流水帳，順手把剛剛拿到的 userID 塞進去！
	err = repository.CreateOrderRecord(tx, req.ProductID, req.Quantity , userID)
	if err != nil {
		fmt.Println("❌ PostgreSQL 拒絕寫入訂單，原因:", err)
		database.RedisClient.IncrBy(c.Request.Context(), "product:1:stock", int64(req.Quantity))
		tx.Rollback() // 資料庫回滾
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "建立訂單明細失敗"})
		return
	}

	// 3. 兩件事都做成功了，才一起 Commit 提交！
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


