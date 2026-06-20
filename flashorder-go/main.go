package main

import (
	"flashorder-go/database" // 引入初始化模組
	"flashorder-go/handler"  // 引入外場路由模組
	"flashorder-go/middleware" // 引入辯識模組
	"github.com/gin-gonic/gin"
	
	//260619 新增防止shutdown、伺服器崩潰導致資料寫入不完全
	"context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
	// 1. 初始化 SQLite 總帳本
	database.InitDB()
	defer database.DB.Close()

	// 2. 初始化 Redis 前線護盾
	database.InitRedis()

	r := gin.Default()

	// 🌟當使用者瀏覽網頁首頁 (http://localhost:8080/) 時，直接送出 public/index.html
	r.StaticFile("/", "./public/index.html")

	// 🌟提供給前端獲取最新庫存的 API
	r.GET("/api/v1/products/1/stock", handler.GetStock)

	// 🌟在 handler.CreateOrder 前面，橫著插一支 middleware.AuthMiddleware()！
	// 這樣請求來的時候，會「由左至右」像闖關一樣，先被警衛檢查，通過了才准進 CreateOrder
	r.POST("/api/v1/orders", middleware.AuthMiddleware(), handler.CreateOrder)

	// 🌟身分驗證大廳
	r.POST("/api/v1/auth/signup", handler.SignUp) // 註冊櫃檯
	r.POST("/api/v1/auth/login", handler.Login)   // 登入櫃檯

	 // 🌟 1. 建立一個自訂的 HTTP 伺服器物件
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r, // 把 gin 的引擎掛載進來
    }

    // 🌟 2. 把伺服器丟到「背景 (Goroutine)」去啟動，不要卡住主執行緒
    go func() {
        log.Println("🚀 伺服器啟動於 port 8080")
        // ListenAndServe 會一直跑，直到被我們強制關閉才會回傳 ErrServerClosed
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("伺服器啟動失敗: %s\n", err)
        }
    }()

    // 🌟 3. 建立一個「訊號接收器 (Channel)」
    quit := make(chan os.Signal, 1)
    
    // 告訴作業系統：如果有人按 Ctrl+C (SIGINT) 或傳送終止訊號 (SIGTERM)，請通知這個 quit 通道
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // 主程式會卡在這裡「等」，直到 quit 通道收到訊號為止
    <-quit
    log.Println("⚠️ 收到關機訊號，準備進行優雅關機...")

    // 🌟 4. 收到關機訊號！拉下鐵門，並設定 5 秒鐘的「最後通牒倒數計時器」
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel() // 習慣動作：用完 ctx 記得釋放資源

    // 呼叫 Shutdown()，它會拒絕新請求，並等待舊請求處理完畢。如果超過 5 秒就會強制中斷。
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("❌ 伺服器關機發生錯誤:", err)
    }

    log.Println("✅ 伺服器已安全、優雅地完全關閉。你可以安心下班了！")
}
	