⚡ FlashOrder 閃電點餐系統 (High-Concurrency Ordering API)

FlashOrder 是一個專為「高併發搶購 / 秒殺情境」設計的後端 API 系統。
本專案拋棄了傳統的單體式義大利麵程式碼，採用 Clean Architecture (乾淨架構) 進行模組化設計，並深度利用 Go 語言的併發特性與 Redis 記憶體資料庫，確保系統在海量流量下依然能穩定運作。

🛠️ 技術棧 (Tech Stack)

後端語言: Golang (Go 1.20+)

Web 框架: Gin

資料庫 / 快取: Redis (go-redis/v8)

架構設計: Clean Architecture (Handler -> Service -> Repository), Dependency Injection (依賴注入)

資安防護: JWT (JSON Web Token), Bcrypt (密碼雜湊)

併發處理: Goroutine, Buffered Channel

🌟 核心架構與亮點 (Architectural Highlights)

1. 乾淨架構與依賴注入 (Clean Architecture & DI)

全系統嚴格遵守三層架構設計，將「HTTP 路由解析 (Handler)」、「商業邏輯 (Service)」與「資料庫操作 (Repository)」徹底解耦。
透過介面 (Interface) 實作依賴注入，不僅讓各模塊的職責單一 (Single Responsibility)，更讓底層資料庫的抽換（例如從 Redis 遷移至 PostgreSQL）零痛感，完全不影響上層業務邏輯。

2. 高併發削峰填谷 (Peak Shaving with Message Queue)

針對「秒殺點餐」情境，放棄傳統的「同步寫入資料庫」作法。
巧妙運用 Go 內建的 Buffered Channel 實作輕量級 Message Queue。主執行緒接收訂單後 0.01 秒即將資料丟入 Channel 並回應客戶端；背景由多個 Goroutines (Worker Pool) 以非同步、平緩的速度將訂單消化並寫入資料庫。完美保護資料庫免於瞬間流量衝擊 (DB Crash)。

3. API 雙重防禦網 (Rate Limiting & JWT Auth)

第一道防線 (限流防 DDoS)：基於 Redis 實作「固定時間窗限流器 (Rate Limiter)」，有效阻擋惡意腳本連點，保障正常用戶權益。

第二道防線 (身分驗證)：自訂 Auth Middleware 攔截並解析 JWT Bearer Token，確保只有合法登入之會員可執行點餐操作。

4. 系統穩定性保護 (Graceful Shutdown)

導入 os/signal 監聽系統中斷訊號 (SIGINT/SIGTERM)。當伺服器需要更新或關閉時，系統會拒絕新的連線，並給予背景 Worker 緩衝時間處理完手中現有的訂單，確保「零掉單、零狀態異常」，實現優雅關機。

🚀 API 列表 (Endpoints)

方法

路由

功能

防護要求

請求參數範例

POST

/register

會員註冊

無

{"username":"mike", "password":"123"}

POST

/order?id=xxx

送出點餐

限流 + JWT 驗證

無 (依賴 JWT Header 與 URL Query)

(註：登入配發 Token 之 /login 功能即將上線，目前系統保留了企業級的 AuthService 以隨時擴充發放邏輯)

💻 如何在本地端啟動 (How to Run)

確保本機已安裝 Go 與啟動 Redis 伺服器 (Port: 6379)。

Clone 此專案：

git clone [https://github.com/你的帳號/flashorder-go.git](https://github.com/你的帳號/flashorder-go.git)
cd flashorder-go


下載相依套件：

go mod tidy


啟動伺服器：

go run main.go
