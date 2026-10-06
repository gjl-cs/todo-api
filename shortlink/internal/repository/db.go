package repository

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"os"
	"time"
)

func NewDB() (*sql.DB, error) {
	config := mysql.Config{
		User:   os.Getenv("DB_USER"),
		Passwd: os.Getenv("DB_PASSWORD"),
		Net:    "tcp",
		Addr: fmt.Sprintf(
			"%s:%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
		),
		DBName:    os.Getenv("DB_NAME"),
		ParseTime: true,
	}
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
