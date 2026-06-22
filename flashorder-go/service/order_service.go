package service

import (
	"errors"
	"log"

    // 💡 這裡要引入你剛剛寫好的 repository 模組
    // (注意：這裡的 flashorder-go 要替換成你 go.mod 裡面的 module 名稱)
	"flashorder-go/repository"
)

// ==========================================
// 1. 定義介面 —— 這是主廚的「合約」
// ==========================================
type OrderService interface {
	ProcessOrder(orderID string) error
}

// ==========================================
// 2. 定義實體結構 —— 這是主廚「本人」
// ==========================================
type orderServiceImpl struct {
	// 🌟 核心關鍵：主廚的口袋裡，隨時帶著一個「倉管員 (Repository)」
    // 注意這裡我們依賴的是「介面 (Interface)」，而不是某個具體的實作！
	repo repository.OrderRepository
}

// ==========================================
// 3. 工廠函式 —— 聘請主廚，並把倉管員分配給他
// ==========================================
func NewOrderService(r repository.OrderRepository) OrderService {
	return &orderServiceImpl{
		repo: r, // 把傳進來的倉管員，收進主廚的口袋裡
	}
}

// ==========================================
// 4. 實作合約方法 —— 主廚開始做菜 (商業邏輯)
// ==========================================
func (s *orderServiceImpl) ProcessOrder(orderID string) error {
	log.Printf("👨‍🍳 [Service] 收到訂單請求，開始驗證邏輯... 訂單號: %s\n", orderID)

	// 💡 商業邏輯 1：檢查訂單號碼是否為空
	if orderID == "" {
		log.Println("❌ [Service] 錯誤：訂單號碼不能為空！")
		return errors.New("invalid order ID")
	}

	// 💡 商業邏輯 2：(未來可以在這裡加入 Redis 扣庫存的邏輯)
	log.Println("✅ [Service] 驗證通過！準備交給 Repository 寫入資料庫...")

	// 🌟 呼叫口袋裡的倉管員，幫忙把資料存進去
	err := s.repo.SaveOrder(orderID)
	if err != nil {
		return err
	}

	return nil
}
