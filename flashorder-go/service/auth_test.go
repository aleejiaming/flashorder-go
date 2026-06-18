package service

import (
	"testing"
)

// 測試目標：檢查 HashPassword 與 CheckPassword 是否能完美配合
func TestPasswordHashing(t *testing.T) {
	// 🟢 1. Arrange (準備階段)：準備我們要測試的假資料
	plainPassword := "MySuperSecret123!"

	// 🟢 2. Act (執行階段)：呼叫我們真正寫的商務邏輯
	// 假設你在 service 裡有寫這兩個函式：HashPassword 和 CheckPassword
	hashedPassword, err := HashPassword(plainPassword)

	// 🟢 3. Assert (驗證階段)：裁判逼逼！檢查結果對不對
	// 驗證 A：加密過程中絕對不能出錯
	if err != nil {
		t.Fatalf("❌ 加密失敗，預期沒有錯誤，卻得到: %v", err)
	}

	// 驗證 B：加密出來的字串絕對不能跟明文一模一樣
	if hashedPassword == plainPassword {
		t.Errorf("❌ 嚴重漏洞：密碼根本沒有被加密！")
	}

	// 驗證 C：拿原本的密碼去核對，必須能解鎖成功
	isValid := CheckPasswordHash(plainPassword, hashedPassword)
	if !isValid {
		t.Errorf("❌ 驗證失敗：正確的密碼無法解開自己產生的 Hash")
	}

	// 驗證 D：拿錯誤的密碼去核對，必須被擋下來
	isInvalid := CheckPasswordHash("WrongPassword", hashedPassword)
	if isInvalid {
		t.Errorf("❌ 嚴重漏洞：隨便亂打的密碼竟然驗證成功了！")
	}
}