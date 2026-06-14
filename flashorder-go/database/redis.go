package database

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// RDB 變數大寫，讓全專案的 handler 都能呼叫它
var RDB *redis.Client

// Go 語言調用 Redis 時，習慣需要帶一個上下文 (Context) 用來控制超時
var Ctx = context.Background()

func InitRedis() {
	// 1. 連線到本地的 Redis 伺服器
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 預設地址
		Password: "",               // 預設沒有密碼
		DB:       0,                // 使用預設的第 0 個資料庫
	})

	// 2. 發送一個 Ping 測試連線是否真的通了
	pong, err := RDB.Ping(Ctx).Result()
	if err != nil {
		panic("Redis 連線失敗: " + err.Error())
	}

	fmt.Printf("🚀 Redis 連線成功！伺服器回應: %s\n", pong)

	// 3. 搶先在 Redis 裡初始化我們的「招牌牛腩」活動庫存為 5 份
	// Key 叫做 "product:1:stock"
	err = RDB.Set(Ctx, "product:1:stock", 5, 0).Err()
	if err != nil {
		panic("Redis 初始化庫存失敗: " + err.Error())
	}
	fmt.Println("🔥 [Redis 預熱完畢] 商品 ID 1 庫存 5 份已寫入記憶體！")
}