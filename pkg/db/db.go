// Package db 提供数据库连接管理和通用操作。
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New 创建数据库连接
func New(ctx context.Context, cfg Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("database connected", "host", cfg.Host, "database", cfg.Database)
	return &Database{pool: pool}, nil
}

// Pool 返回连接池
func (d *Database) Pool() *pgxpool.Pool {
	return d.pool
}

// Close 关闭连接
func (d *Database) Close() {
	d.pool.Close()
	slog.Info("database connection closed")
}

// Exec 执行 SQL 语句
func (d *Database) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := d.pool.Exec(ctx, sql, args...)
	return err
}

// QueryRow 查询单行
func (d *Database) QueryRow(ctx context.Context, sql string, args ...any) RowScanner {
	return d.pool.QueryRow(ctx, sql, args...)
}

// Query 查询多行
func (d *Database) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return d.pool.Query(ctx, sql, args...)
}


