Write-Host "🚀 [FlashOrder] 準備啟動開發環境..." -ForegroundColor Cyan

# ==========================================
# 1. 自動啟動 Redis 模組
# ==========================================
Write-Host "📦 正在檢查並喚醒 Redis..." -ForegroundColor Yellow

# 👉 請根據你的電腦情況，把下面「其中一種」方式前面的 '#' 刪掉：

# 方式 A: 如果你是直接下載 Windows 版的 redis-server.exe (請換成你的真實路徑)
Start-Process -FilePath "C:\C:\Users\MingLee\Desktop\sideproject\infrastructure\Redis\redis-server.exe" -WindowStyle Minimized

# 方式 B: 如果你是用 WSL (Windows 裡的 Linux) 啟動 Redis
# wsl service redis-server start

# 方式 C: 如果你是用 Docker 跑 Redis (業界最常用)
# docker start flashorder-redis

# 給 Redis 2 秒鐘的時間開機暖身
Start-Sleep -Seconds 2

# ==========================================
# 2. 啟動 Go 伺服器
# ==========================================
Write-Host "🔥 啟動 Go 伺服器..." -ForegroundColor Green
cd flashorder-go
go run main.go