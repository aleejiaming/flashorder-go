package database

import (
	"database/sql"
	"fmt"
	"time"

	// 🌟 引入全新的 PostgreSQL 驅動
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB() {
	// 🚨【請根據你的設定修改】帳號預設 postgres，密碼請輸入你剛剛在安裝精靈設定的密碼！
	// 格式：postgres://帳號:密碼@主機位置:埠號/資料庫名稱?參數
	dsn := "postgres://postgres:Ming741852@localhost:5432/flashorder?sslmode=disable"

	var err error
	// 🌟 驅動名稱換成 "pgx"
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		panic("資料庫連線失敗: " + err.Error())
	}

	// 設定連線池
	// 💡 既然換成了正規 PostgreSQL，我們再也不用憋屈地開 MaxOpenConns(1) 了！
	// 這裡直接開到 10，體驗正規伺服器的並發吞吐量！
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// 測試連線是否真的通了
	err = DB.Ping()
	if err != nil {
		panic("資料庫 Ping 失敗，請檢查密碼或連線字串: " + err.Error())
	}

	fmt.Println("💾 [PostgreSQL 總帳本] 連線成功！實體行級鎖引擎就位。")
}