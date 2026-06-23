package service

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ==========================================
// 1. 定義介面 —— 保安主管的合約
// ==========================================
type AuthService interface {
	ValidateToken(tokenString string) (string, error) 
	// 💡 未來你也可以在這裡加入 GenerateToken(userID string) (string, error) 來負責登入發派 Token
}

// ==========================================
// 2. 定義實體 —— 保安主管本人
// ==========================================
type authServiceImpl struct {
	secretKey []byte // 主管的專屬解密金鑰
}

// ==========================================
// 3. 工廠函式 —— 聘請保安主管
// ==========================================
func NewAuthService(secret string) AuthService {
	return &authServiceImpl{
		secretKey: []byte(secret),
	}
}

// ==========================================
// 4. 實作技能 —— 真槍實彈的 JWT 解密邏輯
// ==========================================
func (s *authServiceImpl) ValidateToken(tokenString string) (string, error) {
	// 解析並驗證 Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 確保加密演算法是我們預期的 HMAC (防禦安全漏洞)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非預期的簽名方法: %v", token.Header["alg"])
		}
		// 交出金鑰給套件去解密
		return s.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	// 驗證 Token 是否合法，並把裡面的 Payload (Claims) 拿出來
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 假設我們發放 Token 時，是把使用者 ID 存在 "user_id" 這個欄位
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("token 中找不到 user_id 欄位")
		}
		return userID, nil
	}

	return "", errors.New("無效的 token")
}