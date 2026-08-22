// Package reporter 提供报告生成功能。
package reporter

import (
	"context"

	"aitrendings/core/model"
)

// Store 数据访问接口
type Store interface {
	GetRecentSnapshots(ctx context.Context, days int) ([]*model.Snapshot, error)
}

// ComparisonReporter 横向对比报告生成器
type ComparisonReporter struct {
	store Store
}
