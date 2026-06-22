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

	// 💡 引入我們剛剛寫好的三層架構套件
	// 注意：請確保 flashorder-go 符合你 go.mod 裡面的 module 名稱
	"flashorder-go/handler"
	"flashorder-go/repository"
	"flashorder-go/service"
)

func main() {
	// 1. 初始化 Gin 路由器 (準備好餐廳的場地)
	r := gin.Default()

	// ==========================================
	// 🌟 核心組裝區：依賴注入 (Dependency Injection)
	// ==========================================
	// 這裡的順序絕對不能亂！因為外層依賴內層，所以我們必須「由內而外」組裝。

	// 步驟 A：聘請一個倉管員 (最內層)
	// (未來這裡會傳入真實的 db 連線，例如 repository.NewOrderRepository(db))
	orderRepo := repository.NewOrderRepository()

	// 步驟 B：聘請一個主廚，並把倉管員配發給他
	orderService := service.NewOrderService(orderRepo)

	// 步驟 C：聘請一個外場服務生，並把主廚配發給他 (最外層)
	orderHandler := handler.NewOrderHandler(orderService)

	// ==========================================
	// 📍 路由綁定 (Routing)
	// ==========================================
	// 告訴路由器：如果有人發送 POST 到 /order，就請 orderHandler 服務生去接待他
	r.POST("/order", orderHandler.CreateOrder)


	// ==========================================
	// 🛑 啟動伺服器與「優雅關機」保護傘
	// ==========================================
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 把伺服器丟到背景去跑
	go func() {
		log.Println("🚀 閃電點餐系統 (Clean Architecture 版本) 啟動於 port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("伺服器啟動失敗: %s\n", err)
		}
	}()

	// 建立中斷訊號接收器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 卡在這裡等關機訊號

	log.Println("⚠️ 收到關機訊號，準備進行優雅關機...")

	// 給予 5 秒鐘的緩衝時間，把手上的訂單處理完
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ 伺服器關機錯誤:", err)
	}

	log.Println("✅ 伺服器已安全關閉。")
}