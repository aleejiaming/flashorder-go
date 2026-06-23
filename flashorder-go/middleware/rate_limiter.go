package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimiter 是一個閉包工廠，用來產生限流的 Middleware
// rdb: 傳入我們在 main 建立好的 Redis 連線
// limit: 設定一個時間內最多可以發送幾次請求 (例如: 3 次)
// window: 設定這個時間窗口有多大 (例如: 1 * time.Second 代表 1 秒內)
func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// 1. 決定「誰」是限流的對象？
		// 這裡我們簡單用客人的 IP 位址當作辨識碼 (實務上可能會用 JWT 裡的 UserID)
		clientIP := c.ClientIP()

		// 為了讓 Redis 裡的 Key 看起來整齊，我們加個前綴
		redisKey := fmt.Sprintf("rate_limit:%s", clientIP)

		// 2. 向 Redis 詢問：這個 IP 目前點了幾次？
		// INCR 指令很神奇，如果 Key 不存在，它會先幫你建一個 0，然後 +1 變成 1。
		// 如果 Key 存在，它就直接把數字 +1 並回傳。
		currentCount, err := rdb.Incr(ctx, redisKey).Result()
		if err != nil {
			// 如果 Redis 當機，為了不影響客人點餐，我們選擇「放行」
			// (實務上可以依據公司政策決定要不要擋)
			fmt.Println("⚠️ [限流器] Redis 連線異常，預設放行:", err)
			c.Next()
			return
		}

		// 3. 如果這是他這個時間窗內的「第一次」點擊，我們要幫這個 Key 設定「過期時間」
		// 這樣 1 秒後，這個計數器就會消失重新計算。
		if currentCount == 1 {
			rdb.Expire(ctx, redisKey, window)
		}

		// 4. 關鍵判斷：他是不是點太快了？
		if int(currentCount) > limit {
			// 警衛嗶嗶！超過限制次數！
			fmt.Printf("🛑 [限流器] 擋下惡意請求！ IP: %s 已在一秒內點擊 %d 次\n", clientIP, currentCount)

			// 直接回傳 HTTP 429 Too Many Requests，並中斷請求
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "點太快啦！",
				"message": "系統正忙，請稍後再試。",
			})
			c.Abort() // ❗重要：這會阻止請求繼續往下走到 Handler (服務生) 那裡
			return
		}

		// 5. 檢查通過，警衛放行！
		// c.Next() 會把請求交給下一個處理器 (也就是我們的 orderHandler)
		c.Next()
	}
}