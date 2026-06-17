package repository

import (
	"database/sql"
)

// User 結構體，用來對應資料庫的欄位
type User struct {
	ID           int
	Username     string
	PasswordHash string
}

// CreateUser 實體寫入一個新會員
func CreateUser(db *sql.DB, username, passwordHash string) error {
	_, err := db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, passwordHash)
	return err
}

// GetUserByUsername 透過帳號尋找會員（登入比對時使用）
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}