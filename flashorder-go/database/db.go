package database

import (
	"database/sql"
	"fmt"
	_ "github.com/glebarez/go-sqlite"
)

// DB 變數大寫，允許跨套件（Package）存取
var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "orders.db")
	if err != nil {
		panic("資料庫連線失敗: " + err.Error())
	}

	// 🔓 解開連線封印！允許同時開啟 10 條連線，讓讀取速度飆升！
	DB.SetMaxOpenConns(10)
	// 資料表升級：加入 version 欄位
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		stock INTEGER NOT NULL,
		version INTEGER NOT NULL DEFAULT 0
	);`
	_, err = DB.Exec(createTableSQL)
	if err != nil {
		panic("建立資料表失敗: " + err.Error())
	}

	// 初始化初始資料
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM products WHERE id = 1").Scan(&count)
	if err != nil {
		panic("檢查初始資料失敗: " + err.Error())
	}

	if count == 0 {
		_, err = DB.Exec("INSERT INTO products (id, name, stock) VALUES (1, '招牌牛腩', 5)")
		if err != nil {
			panic("初始化商品失敗: " + err.Error())
		}
		fmt.Println("🎉 資料庫初始化成功！已成功建立『招牌牛腩』庫存：5 份")
	}
}