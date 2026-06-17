package main

import (
	"flashorder-go/database" // 引入初始化模組
	"flashorder-go/handler"  // 引入外場路由模組
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化 SQLite 總帳本
	database.InitDB()
	defer database.DB.Close()

	// 2. 初始化 Redis 前線護盾
	database.InitRedis()

	r := gin.Default()

	// 🌟【新增】當使用者瀏覽網頁首頁 (http://localhost:8080/) 時，直接送出 public/index.html
	r.StaticFile("/", "./public/index.html")

	// 🌟【新增】提供給前端獲取最新庫存的 API
	r.GET("/api/v1/products/1/stock", handler.GetStock)

	// 3. 原有的秒殺下單 API
	r.POST("/api/v1/orders", handler.CreateOrder)

	// 🌟【全新加入】身分驗證大廳
	r.POST("/api/v1/auth/signup", handler.SignUp) // 註冊櫃檯
	r.POST("/api/v1/auth/login", handler.Login)   // 登入櫃檯

	r.Run(":8080")
}