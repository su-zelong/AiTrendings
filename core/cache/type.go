// Package cache 提供 LLM 分析结果缓存功能。
package cache

import (
	"context"
	"time"

	"aitrendings/core/model"
)

// Store 缓存存储接口
type Store interface {
	GetAnalysisCache(ctx context.Context, readmeMD5 string) (*model.CachedAnalysis, error)
	SetAnalysisCache(ctx context.Context, readmeMD5 string, response *model.LLMResponse, ttl time.Duration) error
}

// ReadmeCache README 分析缓存
type ReadmeCache struct {
	store Store
	ttl   time.Duration
}
