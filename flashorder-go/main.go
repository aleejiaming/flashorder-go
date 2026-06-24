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

	// 🚨 這裡的 flashorder-go 請確認與你 go.mod 裡面的一致
	"flashorder-go/handler"
	"flashorder-go/middleware"
	"flashorder-go/repository"
	"flashorder-go/service"
)

// ==========================================
// 👷 建立「打工人」的工作邏輯 (專門處理點餐排隊)
// ==========================================
func worker(workerID int, queue chan string, repo repository.OrderRepository) {
	log.Printf("👷 [打工人 %d] 已上線，死死盯著排隊箱...\n", workerID)

	for {
		orderID := <-queue
		log.Printf("👷 [打工人 %d] 從箱子拿出訂單 %s，開始慢慢處理...\n", workerID, orderID)
		
		time.Sleep(2 * time.Second) // 模擬處理時間
		
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
	// 🌟 核心組裝區：依賴注入 (Dependency Injection)
	// ==========================================

	// --- 🔒 共用服務 ---
	// 聘請保安主管，給他一把只有餐廳知道的「解密金鑰」
	authService := service.NewAuthService("flashorder-super-secret-key")

	// --- 👤 會員模塊 (User Module) 生產線 ---
	userRepo := repository.NewRedisUserRepository(rdb)   // 1. 聘請會員倉管員
	userService := service.NewUserService(userRepo)      // 2. 聘請會員主廚
	userHandler := handler.NewUserHandler(userService)   // 3. 聘請會員服務生

	// --- 🍔 點餐模塊 (Order Module) 生產線 ---
	orderQueue := make(chan string, 1000)                // 1. 買一個排隊箱
	orderRepo := repository.NewRedisOrderRepository(rdb) // 2. 聘請點餐倉管員
	for i := 1; i <= 3; i++ {                            // 3. 啟動 3 個背景打工人
		go worker(i, orderQueue, orderRepo)
	}
	orderService := service.NewOrderService(orderQueue)  // 4. 聘請點餐主廚 (交給他排隊箱)
	orderHandler := handler.NewOrderHandler(orderService)// 5. 聘請點餐服務生

	// ==========================================
	// 📍 路由綁定 (Routing) - 掛上營業招牌
	// ==========================================

	// 🆕 新增：會員註冊櫃台
	// 客人遞 JSON 信封來註冊。因為還沒登入，所以「不需要」大門警衛攔阻！
	r.POST("/register", userHandler.RegisterUser)

	// 🍔 點餐櫃台 (雙重防護網上線！)
	r.POST("/order",
		middleware.RateLimiter(rdb, 3, 1*time.Second), // 第一道：防連點警衛
		middleware.AuthMiddleware(authService),        // 第二道：驗證手環警衛
		orderHandler.CreateOrder,                      // 最後才交給服務生
	)

	// ==========================================
	// 🛑 啟動伺服器與「優雅關機」保護傘
	// ==========================================
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("🚀 閃電點餐系統 (會員+排隊版) 啟動於 port 8080")
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