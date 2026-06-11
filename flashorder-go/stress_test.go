package main

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	
)

// 這是 Go 內建的測試功能，用來模擬並發
func TestHighConcurrencyOrder(t *testing.T) {
	url := "http://127.0.0.1:8080/api/v1/orders"
	// 模擬點餐的 JSON 資料：買商品 ID 1，數量 1 份
	jsonData := []byte(`{"product_id": 1, "quantity": 1}`)

	const concurrencyCount = 50 // 同時有 50 個人搶單
	var wg sync.WaitGroup

	t.Log("=== 壓力測試開始：50 個併發請求同時衝進伺服器 ===")

	for i := 0; i < concurrencyCount; i++ {
		wg.Add(1)
		// 啟動 Goroutine (類似超輕量執行緒)，發送非同步請求
		go func(userId int) {
			defer wg.Done()

			// 發送 POST 請求
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Error("連線失敗原因:", err) // <--- 加上這行，把隱藏的錯誤印出來！
				return
			}
			defer resp.Body.Close()
		}(i)
	}

	// 等到 50 個請求全部發送並回應完畢
	wg.Wait()
	t.Log("=== 壓力測試結束 ===")
}