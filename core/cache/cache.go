// Package cache 提供 LLM 分析结果缓存功能。
package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"log/slog"
	"time"

	"aitrendings/core/model"
)

// New 创建 ReadmeCache
func New(store Store) *ReadmeCache {
	return &ReadmeCache{
		store: store,
		ttl:   30 * 24 * time.Hour, // 30天
	}
}

// Get 获取缓存的 LLM 响应
func (c *ReadmeCache) Get(ctx context.Context, readme string) (*model.LLMResponse, error) {
	hash := md5.Sum([]byte(readme))
	readmeMD5 := fmt.Sprintf("%x", hash)

	cached, err := c.store.GetAnalysisCache(ctx, readmeMD5)
	if err != nil {
		slog.Info("cache miss", "md5", readmeMD5)
		return nil, nil
	}

	slog.Info("cache hit", "md5", readmeMD5)
	return cached.LLMResponse, nil
}

// Set 设置 LLM 响应缓存
func (c *ReadmeCache) Set(ctx context.Context, readme string, response *model.LLMResponse) error {
	hash := md5.Sum([]byte(readme))
	readmeMD5 := fmt.Sprintf("%x", hash)

	return c.store.SetAnalysisCache(ctx, readmeMD5, response, c.ttl)
}

// ComputeMD5 计算 MD5
func ComputeMD5(readme string) string {
	hash := md5.Sum([]byte(readme))
	return fmt.Sprintf("%x", hash)
}
