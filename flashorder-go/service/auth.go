package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// 🚨【研發重大機密】這是用來幫 JWT 防偽簽名的秘密鑰匙
// 在真實商業環境中，這個 SecretKey 通常會藏在 .env 環境變數裡，絕對不能外流！
var jwtSecret = []byte("FlashOrderGo_Super_Secret_Key_2026")

// CustomClaims 定義我們想塞進 JWT 手環裡的乘客資料
type CustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// =========================================================================
// 🔒 第一部分：Bcrypt 密碼粉碎工廠 (註冊與登入使用)
// =========================================================================

// HashPassword 將客人的明文密碼（如 123456）轉成不可逆的超強大亂碼
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost 代表加密強度（預設是 10 次方算力迭代），能有效防禦暴力破解
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 檢查客人登入時輸入的密碼，跟我們資料庫裡的亂碼是否吻合
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // 如果回傳 nil 代表密碼正確，過關！
}

// =========================================================================
// 🎟️ 第二部分：JWT 防偽手環編織與檢查工廠
// =========================================================================

// GenerateToken 登入成功後，幫會員編織一張專屬的 JWT 數位通行證
func GenerateToken(userID int, username string) (string, error) {
	// 設定通行證內容與過期時間（這裡設定 2 小時後失效）
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 使用 HS256 演算法與我們的「祕密鑰匙」進行數學簽名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken 大門警衛專用：當客人亮出 JWT 手環時，負責驗算真偽並拆解出會員 ID
func ValidateToken(tokenString string) (*CustomClaims, error) {
	// 解析手環
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 檢查手環是否合法、且還在有效期限內
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("無效的通行證")
}