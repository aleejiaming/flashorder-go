package handler

import (
	"net/http"

	// 引入 gin 套件來處理 HTTP 請求
	"github.com/gin-gonic/gin"
	
	// 💡 引入你剛剛寫好的 service 模組
    // (注意：這裡的 flashorder-go 要替換成你 go.mod 裡面的 module 名稱)
	"flashorder-go/service"
)

// ==========================================
// 1. 定義介面 —— 這是服務生的「合約」
// ==========================================
// 注意：Handler 的介面通常會對應到你的 API 路由設計
type OrderHandler interface {
	CreateOrder(c *gin.Context)
}

// ==========================================
// 2. 定義實體結構 —— 這是服務生「本人」
// ==========================================
type orderHandlerImpl struct {
	// 🌟 核心關鍵：服務生的口袋裡，隨時帶著一個「主廚 (Service)」
	// 同樣地，這裡依賴的是「介面 (Interface)」，服務生不挑主廚，只要會做菜就好
	orderService service.OrderService
}

// ==========================================
// 3. 工廠函式 —— 聘請服務生，並把主廚分配給他
// ==========================================
func NewOrderHandler(s service.OrderService) OrderHandler {
	return &orderHandlerImpl{
		orderService: s, // 把傳進來的主廚，分配給這個服務生
	}
}

// ==========================================
// 4. 實作合約方法 —— 服務生開始接待客人 (處理 HTTP 請求)
// ==========================================
func (h *orderHandlerImpl) CreateOrder(c *gin.Context) {
	// 1. 取得客人傳來的訂單號碼 (這裡假設是從 URL 參數拿，例如 /order?id=123)
	orderID := c.Query("id")

	// 2. 把客人的需求交給口袋裡的主廚 (Service) 去處理
	// 服務生不需要知道主廚怎麼做菜，也不需要知道有沒有連資料庫
	err := h.orderService.ProcessOrder(orderID)

	// 3. 根據主廚的回報，決定要跟客人說什麼
	if err != nil {
		// 主廚說失敗了 (例如訂單號為空)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "點餐失敗",
			"message": err.Error(),
		})
		return
	}

	// 主廚說成功了
	c.JSON(http.StatusOK, gin.H{
		"message":  "🎉 點餐成功！",
		"order_id": orderID,
	})
}