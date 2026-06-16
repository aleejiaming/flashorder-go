package repository

import (
	"database/sql"
)

// GetStockForUpdate 查詢當前實體庫存 (這裡我們加入了 FOR UPDATE 行級鎖鎖定)
func GetStockForUpdate(tx *sql.Tx, productID int) (int, error) {
	var stock int
	// 🌟 1. 改用 $1 佔位符
	// 🌟 2. 後方加入 FOR UPDATE，讓 Postgres 在實體資料庫層面幫這 5 個決賽圈的人精準排隊！
	err := tx.QueryRow("SELECT stock FROM products WHERE id = $1 FOR UPDATE", productID).Scan(&stock)
	return stock, err
}

// UpdateStock 更新實體庫存
func UpdateStock(tx *sql.Tx, productID int, newStock int) error {
	// 🌟 改用 $1 和 $2 佔位符
	_, err := tx.Exec("UPDATE products SET stock = $1 WHERE id = $2", newStock, productID)
	return err
}

// =========================================================================
// 舊的樂觀鎖留念（可不改，但若想保持語法正確，一樣要把問號改成 $1 $2 $3）
// =========================================================================
func GetStockAndVersion(tx *sql.Tx, productID int) (int, int, error) {
	var stock, version int
	err := tx.QueryRow("SELECT stock, version FROM products WHERE id = $1", productID).Scan(&stock, &version)
	return stock, version, err
}

func UpdateStockOptimistic(tx *sql.Tx, productID int, newStock int, oldVersion int) (int64, error) {
	res, err := tx.Exec("UPDATE products SET stock = $1, version = version + 1 WHERE id = $2 AND version = $3", newStock, productID, oldVersion)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

	// CreateOrderRecord 🌟 新增：在 PostgreSQL 中寫入一筆真實的訂單流水帳
func CreateOrderRecord(tx *sql.Tx, productID int, quantity int) error {
	// 記得 PostgreSQL 的佔位符是 $1, $2
	_, err := tx.Exec("INSERT INTO orders (product_id, quantity) VALUES ($1, $2)", productID, quantity)
	return err
	}