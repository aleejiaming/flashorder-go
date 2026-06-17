package handler

import (
	"net/http"

	"flashorder-go/database"
	"flashorder-go/repository"
	"flashorder-go/service"

	"github.com/gin-gonic/gin"
)

// AuthRequest 接收前端傳來的帳密 JSON
type AuthRequest struct {
	Username string `json:"username" binding:"required,min=4,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

// SignUp 會員註冊櫃檯
func SignUp(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "欄位格式不符（帳號4-20字，密碼至少6字）"})
		return
	}

	// 🔒 底層邏輯：利用 Bcrypt 將密碼打碎成大亂碼
	hashedPassword, err := service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "密碼加密失敗"})
		return
	}

	// 寫入 PostgreSQL
	err = repository.CreateUser(database.DB, req.Username, hashedPassword)
	if err != nil {
		// 🚨 這裡藏著高併發的底層精髓（稍後解析）
		c.JSON(http.StatusConflict, gin.H{"status": "failed", "message": "帳號已被註冊，或系統繁忙"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "註冊成功！歡迎加入 FlashOrder"})
}

// Login 會員登入櫃檯
func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "請輸入帳號與密碼"})
		return
	}

	// 1. 去資料庫撈出這個使用者的亂碼密碼
	user, err := repository.GetUserByUsername(database.DB, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "帳號或密碼錯誤"})
		return
	}

	// 2. 🧠 比對明文密碼與資料庫的亂碼
	if !service.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failed", "message": "帳號或密碼錯誤"})
		return
	}

	// 3. 🎟️ 比對成功！當場編織一條 JWT 通行證給他
	token, err := service.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "生成通行證失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "登入成功！已核發限時通行證",
		"token":   token, // 👈 前端拿到這串字串後，要自己存在 LocalStorage
	})
}