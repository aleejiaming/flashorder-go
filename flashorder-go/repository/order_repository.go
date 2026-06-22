package repository

import (
	"log"
)

// ==========================================
// 1. 定義介面 (Interface) —— 這是倉管員的「合約」
// ==========================================
// 任何想當 Order 倉管員的人，都必須會實作這份合約裡的所有方法
type OrderRepository interface {
	SaveOrder(orderID string) error
}

// ==========================================
// 2. 定義實體結構 (Struct) —— 這是倉管員「本人」
// ==========================================
// 注意：字首小寫 (orderRepositoryImpl)，代表它是「私有的」，外面的人看不到它
type orderRepositoryImpl struct {
	// 未來我們會在這裡放入真正的資料庫連線，例如：
	// db *gorm.DB
	// redis *redis.Client
}

// ==========================================
// 3. 工廠函式 (Factory) —— 專門「面試並錄取」倉管員的地方
// ==========================================
// 注意：字首大寫 (NewOrderRepository)，代表對外公開。
// 這個工廠的回傳值是「介面 (OrderRepository)」，這是一個極度重要的安全機制！
func NewOrderRepository() OrderRepository {
	return &orderRepositoryImpl{}
}

// ==========================================
// 4. 實作合約方法
// ==========================================
// 讓倉管員本人去執行合約上規定的任務
func (r *orderRepositoryImpl) SaveOrder(orderID string) error {
	// 這裡目前先用印出文字來模擬寫入資料庫
	log.Printf("🗄️ [Repository] 正在將訂單 %s 寫入底層資料庫...\n", orderID)
	return nil
}
