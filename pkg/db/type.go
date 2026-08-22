// Package db 提供数据库连接管理和通用操作。
package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config 数据库配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// Database 数据库连接管理
type Database struct {
	pool *pgxpool.Pool
}

// RowScanner 行扫描器接口
type RowScanner interface {
	Scan(dest ...any) error
}

// Rows 结果集接口
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}
