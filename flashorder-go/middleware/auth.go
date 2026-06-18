package middleware

import (
	"fmt"
	"net/http"
	"strings"

	// 🚨【請根據你的 go.mod 修改路徑】把 sideproject 改成你的模組名稱！
	"flashorder-go/service"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 🌟 大門警衛核心：攔截請求、驗算 JWT 手環
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 從 HTTP Header 裡面撈出名為 "Authorization" 的口袋
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "門口安檢：請先登入取得通行證！"})
			c.Abort() // 🚨【底層邏輯】無情攔截！立刻中斷請求，不讓它往後面的 Handler 走
			return
		}

		// 2. 業界標準格式是 "Bearer <JWT字串>"，我們要用空格把他們切開
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "通行證格式錯誤（必須是 Bearer 開頭）"})
			c.Abort()
			return
		}

		// 3. 🧠 調用我們上次寫的數學工廠，驗算這條手環（parts[1] 就是 JWT 本身）
		claims, err := service.ValidateToken(parts[1])
		if err != nil {
			fmt.Println("❌ JWT 驗算失敗原因:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "無效的通行證或通行證已過期"})
			c.Abort()
			return
		}

		// 4. 🎉 驗算成功！警衛把拆解出來的會員資料，寫入 Gin 的 Context 備忘錄裡
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		// 5. 放行！讓請求高高興興地走進下一個內場 Handler 房間
		c.Next()
	}
}