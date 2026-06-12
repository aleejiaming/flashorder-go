package repository

import (
	"database/sql"
)

// GetStockAndVersion 同時抓取庫存與當前的防偽版本號
func GetStockAndVersion(tx *sql.Tx, productID int) (int, int, error) {
	var stock, version int
	err := tx.QueryRow("SELECT stock, version FROM products WHERE id = ?", productID).Scan(&stock, &version)
	return stock, version, err
}

// UpdateStockOptimistic 樂觀鎖更新：只有當版本號和剛才查到的一模一樣時，才允許寫入，並自動將版本號 + 1
func UpdateStockOptimistic(tx *sql.Tx, productID int, newStock int, oldVersion int) (int64, error) {
	res, err := tx.Exec("UPDATE products SET stock = ?, version = version + 1 WHERE id = ? AND version = ?", newStock, productID, oldVersion)
	if err != nil {
		return 0, err
	}
	
	// RowsAffected 會告訴我們這次 SQL 到底成功修改了幾筆資料
	// 如果是 0，代表版本號對不上，有人偷跑了！
	return res.RowsAffected()
}