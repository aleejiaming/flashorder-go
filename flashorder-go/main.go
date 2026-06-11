package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

type OrderRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

var db *sql.DB

// 🛠️ 新增一個精準診斷工具：用來印出當下連線池的健康檢查報告
func printDBStats(stage string) {
	stats := db.Stats()
	fmt.Printf("📊【連線池動態 | %s】正在使用: %d | 累計排隊次數: %d | 累計排隊總耗時: %v\n",
		stage, stats.InUse, stats.WaitCount, stats.WaitDuration)
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "orders.db")
	if err != nil {
		panic("資料庫連線失敗: " + err.Error())
	}

	// 限制最大連線數為 1
	// 這樣 Go 的連線池就會自動在記憶體裡幫 50 個請求排隊，解決 SQLite 鎖定問題！
	db.SetMaxOpenConns(1)

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		stock INTEGER NOT NULL
	);`
	
	//開啟資料庫交易 (Transaction)
	_, err = db.Exec(createTableSQL)
	if err != nil {
		panic("建立資料表失敗: " + err.Error())
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM products WHERE id = 1").Scan(&count)
	if err != nil {
		panic("檢查初始資料失敗: " + err.Error())
	}

	if count == 0 {
		_, err = db.Exec("INSERT INTO products (id, name, stock) VALUES (1, '招牌牛腩', 5)")
		if err != nil {
			panic("初始化商品失敗: " + err.Error())
		}
		fmt.Println("🎉 資料庫初始化成功！已成功建立『招牌牛腩』庫存：5 份")
	}
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	r.POST("/api/v1/orders", func(c *gin.Context) {
		var req OrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": err.Error()})
			return
		}

		// 🛑 捕捉點一：嘗試獲取連線與開啟交易
		tx, err := db.Begin()
		if err != nil {
			// 💥 精準捕捉：連線池滿了、超時、或中斷的底層原因
			fmt.Printf("❌【🔥 獲取連線失敗】%v\n", err)
			printDBStats("獲取連線階段")
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "伺服器忙碌中"})
			return
		}
		
		// 每次有人成功拿到連線，就偷偷觀察一下後台排隊狀況
		printDBStats("成功拿取連線")

		defer tx.Rollback()

		// 🛑 捕捉點二：查詢庫存階段
		var currentStock int
		err = tx.QueryRow("SELECT stock FROM products WHERE id = ?", req.ProductID).Scan(&currentStock)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "找不到該商品"})
			} else {
				fmt.Printf("❌【🔍 庫存查詢失敗】%v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "系統異常"})
			}
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

		// 模擬真實讀寫延遲
		time.Sleep(50 * time.Millisecond)

		// 🛑 捕捉點三：寫回資料庫階段
		newStock := currentStock - req.Quantity
		_, err = tx.Exec("UPDATE products SET stock = ? WHERE id = ?", newStock, req.ProductID)
		if err != nil {
			fmt.Printf("❌【💾 寫入資料庫失敗】%v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "寫入失敗"})
			return
		}

		// 🛑 捕捉點四：提交交易階段
		err = tx.Commit()
		if err != nil {
			fmt.Printf("❌【🏁 交易提交失敗】%v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "訂單確認失敗"})
			return
		}

		fmt.Printf("【🛡️ DB 安全寫入】商品 ID: %d, 正確扣減後的殘餘庫存為: %d\n", req.ProductID, newStock)

		c.JSON(http.StatusOK, gin.H{
			"status":          "success",
			"message":         "點餐成功，已從資料庫扣除庫存！",
			"remaining_stock": newStock,
		})
	})

	r.Run(":8080")
}