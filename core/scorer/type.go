// Package scorer 提供开发者采纳指数计算功能。
package scorer

import (
	"context"
)

// Store 数据访问接口
type Store interface {
	GetStarDelta(ctx context.Context, projectName string, days int) int
}

// AdoptionScorer 采纳指数计算器
type AdoptionScorer struct {
	store Store
}
