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

	// 🌟 【改進起點】設定最大重試次數，並宣告一些跨迴圈需要的變數
	maxRetries := 3
	var newStock int
	var success bool

	// 🛠️ 開始排隊重試迴圈
	for i := 0; i < maxRetries; i++ {
		
		// 每次重試，都必須重新開啟一個全新、獨立的交易夾子！
		tx, err := database.DB.Begin()
		if err != nil {
			// 如果連開交易都失敗，代表資料庫太忙，我們稍微等 10 毫秒，直接進入下一次 loop 
			time.Sleep(10 * time.Millisecond)
			continue 
		}

		// 1. 重新抓取「當下最新」的庫存與版本號
		currentStock, currentVersion, err := repository.GetStockAndVersion(tx, req.ProductID)
		if err != nil {
			tx.Rollback() // 出事了，單子撕掉
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "找不到該商品"})
				return // 東西都沒賣，提早停損，不用重試
			}
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "查詢出錯"})
			return
		}

		// 檢查庫存是否足夠
		if currentStock < req.Quantity {
			tx.Rollback() // 沒牛腩了，單子撕掉
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status":  "failed",
				"message": fmt.Sprintf("很抱歉，庫存不足。目前剩餘: %d 份", currentStock),
			})
			return // 真的賣完了，直接宣判死刑，不用重試！
		}

		// 模擬真實讀寫延遲 (保留你的 50ms 戰場環境)
		time.Sleep(50 * time.Millisecond)

		// 2. 挑戰寫入資料庫（帶上這一次迴圈查到的最新版本號）
		newStock = currentStock - req.Quantity
		rowsAffected, err := repository.UpdateStockOptimistic(tx, req.ProductID, newStock, currentVersion)
		
		// 🚨【改進核心：碰壁攔截點】
		if err != nil {
			tx.Rollback() // 硬碟互卡鎖定了，這張單作廢
			fmt.Printf("⚠️【第 %d 次嘗試】SQLite 報錯 [%v]，原地等 20ms 後再度重試...\n", i+1, err)
			time.Sleep(20 * time.Millisecond) // 閉眼休息 20ms，錯開並發高峰
			continue // 🔥 關鍵！直接跳過後面，進入下一次 for 迴圈重來！
		}

		if rowsAffected == 0 {
			tx.Rollback() // 版本號被別人偷跑改掉了，單子作廢
			fmt.Printf("⚠️【第 %d 次嘗試】版本號不匹配，重新排隊拿新版本號...\n", i+1)
			time.Sleep(15 * time.Millisecond) // 稍微等一下
			continue // 🔥 關鍵！直接進入下一次 for 迴圈重來！
		}

		// ✨ 走到這裡代表既沒報錯、也成功改到資料了！正式提交！
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			continue
		}

		// 標記成功，並大膽跳出 for 迴圈！
		success = true
		fmt.Printf("【🛡️ 樂觀鎖＋自動重試成功】歷經 %d 次調整，最終扣減成功！殘餘庫存: %d\n", i+1, newStock)
		break
	}

	// 🌟 【改進終點】如果 3 次機會都用完了還是失敗，才對前端吐出挫折訊息
	if !success {
		fmt.Println("❌【⚡ 樂觀鎖戰敗】已自動重試 3 次，系統依然過於擁擠，宣告放棄。")
		c.JSON(http.StatusConflict, gin.H{"status": "failed", "message": "伺服器極度繁忙，請稍後再試。"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"message":         "點餐成功！",
		"remaining_stock": newStock,
	})
}