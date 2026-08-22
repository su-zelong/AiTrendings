// Package http 提供 HTTP 客户端工具。
package http

import (
	"net/http"
	"time"
)

// Config 客户端配置
type Config struct {
	MaxAttempts int
	RetryDelay  time.Duration
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	MaxAttempts: 3,
	RetryDelay:  3 * time.Second,
}

// Client 带重试的 HTTP 客户端
type Client struct {
	client *http.Client
	config Config
}
