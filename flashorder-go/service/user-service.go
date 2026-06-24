package service

import (
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	
	// 🚨 記得把這裡的 flashorder-go 換成你實際的 module 名稱！
	"flashorder-go/repository"
)

// ==========================================
// 1. 零件規格書 (Interface)
// ==========================================
type UserService interface {
	Register(username string, password string) error
}

// ==========================================
// 2. 實體機器 (Struct)
// ==========================================
type userServiceImpl struct {
	// 主廚口袋裡帶著「會員倉管員」的聯絡方式
	userRepo repository.UserRepository
}

// ==========================================
// 3. 工廠函式 (Factory)
// ==========================================
func NewUserService(repo repository.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: repo,
	}
}

// ==========================================
// 4. 機器運作邏輯 (主廚做菜)
// ==========================================
func (s *userServiceImpl) Register(username string, password string) error {
	log.Printf("👨‍🍳 [會員主廚] 開始處理 %s 的註冊請求...\n", username)

	// 1. 呼叫倉管員，檢查有沒有人註冊過了
	exists, err := s.userRepo.CheckUserExists(username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("這個帳號已經被註冊過囉")
	}

	// 2. 啟動密碼攪碎機 (Bcrypt)
	// DefaultCost 是攪碎的強度，數字越大越安全但也越慢
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("❌ [會員主廚] 密碼攪碎機故障:", err)
		return errors.New("系統錯誤，無法處理密碼")
	}
	hashedPassword := string(hashedBytes)

	// 3. 密碼攪碎完成！呼叫倉管員存進金庫
	err = s.userRepo.SaveUser(username, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}