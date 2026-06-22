package service

import (
	"errors"
	"log"

    // 💡 這裡要引入你剛剛寫好的 repository 模組
    // (注意：這裡的 flashorder-go 要替換成你 go.mod 裡面的 module 名稱)
	//"flashorder-go/repository"
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
	// 主廚現在不帶倉管員了，他改帶一個「排隊箱 (Channel)」！
	// chan string 代表這個箱子只能裝字串 (訂單號)
	// <-chan 代表「只能拿出來」，chan<- 代表「只能丟進去」，chan 則是「都可以」
	// 這裡主廚只需把單子丟進去，所以是 chan
	orderQueue chan string 
}

// ==========================================
// 3. 工廠函式 —— 聘請主廚，並把倉管員分配給他
// ==========================================
func NewOrderService(queue chan string) OrderService {
	return &orderServiceImpl{
		orderQueue: queue, // 把排隊箱交給主廚
	}
}


func (s *orderServiceImpl) ProcessOrder(orderID string) error {
	log.Printf("👨‍🍳 [Service] 收到訂單請求，開始驗證邏輯... 訂單號: %s\n", orderID)

	// 檢查訂單號碼
	if orderID == "" {
		return errors.New("invalid order ID")
	}

	
	// ==========================================
	// 4. 實作合約方法 (主廚做菜邏輯)
	// ==========================================
	// 🌟 魔法發生在這裡！
	// 主廚驗證完畢後，不再呼叫 repo.SaveOrder。
	// 他直接把訂單號碼「丟上輸送帶 (排隊箱)」，然後立刻跟服務生說 OK！
	
	// 語法： 輸送帶 <- 物品
	s.orderQueue <- orderID
	
	log.Printf("📥 [Service] 訂單 %s 已快速丟入排隊箱！主廚立刻去接下一單！\n", orderID)

	// 立刻回傳 nil (成功)，完全不浪費時間等資料庫寫入
	return nil
}