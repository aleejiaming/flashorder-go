package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"

	"flashorder-go/handler"
	"flashorder-go/repository"
	"flashorder-go/service"
	"flashorder-go/middleware"
)
// ==========================================
// 👷 建立「打工人」的工作邏輯 (這是全新的函數，寫在 main 的上面)
// ==========================================
// 這個打工人需要：一個箱子 (queue) 拿單子，和一個倉管員 (repo) 存單子。
func worker(workerID int, queue chan string, repo repository.OrderRepository) {
	log.Printf("👷 [打工人 %d] 已上線，死死盯著排隊箱...\n", workerID)
	
	for {
		// 語法： 物品 <- 輸送帶
		// 這裡會「卡住」，直到箱子裡有訂單掉下來為止
		orderID := <-queue 
		
		log.Printf("👷 [打工人 %d] 從箱子拿出訂單 %s，開始慢慢處理...\n", workerID, orderID)
		
		// 模擬真實資料庫寫入很慢的情況 (睡 2 秒)
		time.Sleep(2 * time.Second) 
		
		// 呼叫倉管員去存檔
		err := repo.SaveOrder(orderID)
		if err != nil {
			log.Printf("❌ [打工人 %d] 處理訂單 %s 失敗: %v\n", workerID, orderID, err)
		} else {
			log.Printf("✅ [打工人 %d] 訂單 %s 處理完成！\n", workerID, orderID)
		}
	}
}
func main() {
	// 0. 準備好 Redis 的連線鑰匙
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 1. 初始化 Gin 路由器
	r := gin.Default()

	// ==========================================
	// 🌟 核心組裝區：依賴注入 (加入排隊箱)
	// ==========================================

	// 步驟 1：建立一個長度 1000 的排隊箱 (有緩衝 Channel)
	orderQueue := make(chan string, 1000)

	// 步驟 2：聘請一個真實的 Redis 倉管員
	orderRepo := repository.NewRedisOrderRepository(rdb)

	// 步驟 3：啟動 3 個背景打工人！
	// 使用 go 關鍵字，可以把他們丟到背景獨立執行 (Goroutine)
	for i := 1; i <= 3; i++ {
		go worker(i, orderQueue, orderRepo) 
	}

	//【新增步驟】驗證使用者資訊，回傳一組「解密金鑰」
	authService := service.NewAuthService("flashorder-super-secret-key")

	// 步驟 4：聘請主廚，【注意：現在交給主廚的是排隊箱 orderQueue】
	orderService := service.NewOrderService(orderQueue)

	// 步驟 5：聘請外場服務生
	orderHandler := handler.NewOrderHandler(orderService)

	// ==========================================
	// 🎟️ 臨時發放機：不需要連資料庫，直接給你一張合法的 JWT 通行證
	// ==========================================
	r.GET("/test-login", func(c *gin.Context) {
		// 1. 製造一張手環，把使用者 ID 設為 "VIP_999"
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "VIP_999", 
		})

		// 2. 蓋上跟保安主管一模一樣的「專屬金鑰」印章！
		// (注意：這裡的字串必須跟 authService 裡的一模一樣)
		tokenString, err := token.SignedString([]byte("flashorder-super-secret-key"))
		if err != nil {
			c.JSON(500, gin.H{"error": "手環製作失敗"})
			return
		}

		// 3. 把手環交給客人
		c.JSON(200, gin.H{
			"message": "登入成功！請複製下面的 token",
			"token":   tokenString,
		})
	})

	// ==========================================
	// 📍 路由綁定 (Routing)
	// 📍 路由與啟動 (跟之前一樣，具備優雅關機)
	// ==========================================
	r.POST("/order",
		middleware.RateLimiter(rdb, 3, 1*time.Second), // 第一道：防連點 (每秒最多 3 次)
		middleware.AuthMiddleware(authService),        // 第二道：驗證 JWT 手環 (傳入保安主管)
		orderHandler.CreateOrder,                      // 最後：服務生接單
		)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("🚀 閃電點餐系統 (高併發排隊版) 啟動於 port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("伺服器啟動失敗: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️ 收到關機訊號，準備進行優雅關機...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ 伺服器關機錯誤:", err)
	}
	log.Println("✅ 伺服器已安全關閉。")
}