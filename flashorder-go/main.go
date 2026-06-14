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
	r.POST("/api/v1/orders", handler.CreateOrder)
	r.Run(":8080")
}