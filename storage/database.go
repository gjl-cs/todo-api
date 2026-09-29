package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "todo.db")
	if err != nil {
		panic(err)
	}
	createTable()
	migrateTable()
}
func createTable() {
	sql := `
	CREATE TABLE IF NOT EXISTS todos(
		id INTEGER PRIMARY KEY,
		title TEXT,
		done BOOLEAN,
		priority INTEGER,
		created_at TEXT
	);
	`
	_, err := DB.Exec(sql)
	if err != nil {
		panic(err)
	}
}
func migrateTable() {
	_, err := DB.Exec(`
		ALTER TABLE todos ADD COLUMN created_at TEXT
	`)
	if err != nil {
		fmt.Println("数据库迁移:", err)
		return
	}
	fmt.Println("数据库迁移成功:已添加 due_date 字段")
}