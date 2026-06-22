package repository

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

// ==========================================
// 1. 定義實體結構 —— 這是「Redis 專長」的倉管員
// ==========================================
type redisOrderRepositoryImpl struct {
	// 🌟 關鍵：這個倉管員的口袋裡，裝著一把打開 Redis 倉庫的鑰匙 (Client)
	client *redis.Client
}

// ==========================================
// 2. 工廠函式 —— 聘請他時，必須把鑰匙交給他
// ==========================================
// 注意：他回傳的依然是 OrderRepository (戴著跟假倉管員一樣的面具！)
func NewRedisOrderRepository(c *redis.Client) OrderRepository {
	return &redisOrderRepositoryImpl{
		client: c,
	}
}

// ==========================================
// 3. 實作合約方法 —— 真槍實彈把資料存進 Redis
// ==========================================
func (r *redisOrderRepositoryImpl) SaveOrder(orderID string) error {
	ctx := context.Background()

	// 把訂單號碼存進 Redis，設定值為 "processing" (處理中)，0 代表不設定過期時間
	err := r.client.Set(ctx, "order:"+orderID, "processing", 0).Err()
	if err != nil {
		log.Printf("❌ [Redis Repo] 寫入失敗: %v\n", err)
		return err
	}

	log.Printf("🗄️ [Redis Repo] 成功將訂單 %s 寫入真實的 Redis 資料庫！\n", orderID)
	return nil
}