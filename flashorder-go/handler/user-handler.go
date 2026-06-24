package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	
	// 🚨 記得把這裡的 flashorder-go 換成你實際的 module 名稱！
	"flashorder-go/service"
)

// 💡 預先定義好「JSON 信封」的格式 (原物料清單)
// binding:"required" 意思是如果客人沒填這個欄位，Gin 會直接報錯
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ==========================================
// 1. 零件規格書 (Interface)
// ==========================================
type UserHandler interface {
	RegisterUser(c *gin.Context)
}

// ==========================================
// 2. 實體機器 (Struct)
// ==========================================
type userHandlerImpl struct {
	userService service.UserService // 口袋裝著會員主廚
}

// ==========================================
// 3. 工廠函式 (Factory)
// ==========================================
func NewUserHandler(s service.UserService) UserHandler {
	return &userHandlerImpl{
		userService: s,
	}
}

// ==========================================
// 4. 機器運作邏輯 (櫃台接待)
// ==========================================
func (h *userHandlerImpl) RegisterUser(c *gin.Context) {
	// 1. 準備一個空的托盤來裝客人的信封
	var req RegisterRequest

	// 2. 拆開 JSON 信封 (ShouldBindJSON)
	// 如果信封格式不對，或是少填了必填欄位，就直接退回
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "信封格式錯誤或缺少帳號密碼"})
		return
	}

	// 3. 轉身把拆出來的原物料 (帳號密碼) 交給主廚 (Service)
	err := h.userService.Register(req.Username, req.Password)

	// 4. 根據主廚的回報，端出結果給客人
	if err != nil {
		// 如果失敗 (例如帳號重複)，回傳 400
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 成功註冊！
	c.JSON(http.StatusOK, gin.H{"message": "🎉 註冊成功！歡迎加入閃電點餐！"})
}