// Package http 提供 HTTP 客户端工具。
package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NewClient 创建 HTTP 客户端
func NewClient(client *http.Client, config Config) *Client {
	return &Client{
		client: client,
		config: config,
	}
}

// Do 发起请求（带重试）
func (c *Client) Do(ctx context.Context, method, url string, body io.Reader, contentType string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= c.config.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode == http.StatusOK {
			return respBody, nil
		}

		if !IsRetryable(resp.StatusCode) {
			return nil, fmt.Errorf("http error (status %d): %s", resp.StatusCode, truncate(string(respBody), 200))
		}

		lastErr = fmt.Errorf("http error (status %d): %s", resp.StatusCode, truncate(string(respBody), 200))
		backoff := time.Duration(1<<(attempt-1)) * c.config.RetryDelay
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("retries exhausted: %w", lastErr)
}

// IsRetryable 判断是否可重试
func IsRetryable(statusCode int) bool {
	return statusCode == 429 || (statusCode >= 500 && statusCode < 600)
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
