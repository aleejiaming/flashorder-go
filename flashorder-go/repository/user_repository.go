package repository

import (
	"context"
	"errors"
	"log"

	"github.com/go-redis/redis/v8"
)

// ==========================================
// 1. 零件規格書 (Interface)
// ==========================================
type UserRepository interface {
	CheckUserExists(username string) (bool, error)
	SaveUser(username string, hashedPassword string) error
}

// ==========================================
// 2. 實體機器 (Struct)
// ==========================================
type redisUserRepositoryImpl struct {
	client *redis.Client // 裝載 Redis 倉庫鑰匙
}

// ==========================================
// 3. 工廠函式 (Factory)
// ==========================================
func NewRedisUserRepository(c *redis.Client) UserRepository {
	return &redisUserRepositoryImpl{
		client: c,
	}
}

// ==========================================
// 4. 機器運作邏輯
// ==========================================

// 檢查帳號是否存在
func (r *redisUserRepositoryImpl) CheckUserExists(username string) (bool, error) {
	ctx := context.Background()
	redisKey := "user:" + username

	// Exists 會回傳 1 (存在) 或 0 (不存在)
	count, err := r.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 儲存帳號與雜湊密碼
func (r *redisUserRepositoryImpl) SaveUser(username string, hashedPassword string) error {
	ctx := context.Background()
	redisKey := "user:" + username

	// 將帳號存入 Redis，沒有設定過期時間 (0)
	err := r.client.Set(ctx, redisKey, hashedPassword, 0).Err()
	if err != nil {
		log.Printf("❌ [會員倉管] 儲存會員 %s 失敗: %v\n", username, err)
		return errors.New("資料庫寫入失敗")
	}

	log.Printf("🗄️ [會員倉管] 會員 %s 已安全存入金庫！\n", username)
	return nil
}