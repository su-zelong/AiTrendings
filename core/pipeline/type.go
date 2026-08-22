// Package pipeline 提供核心流水线编排功能。
package pipeline

import (
	"context"

	"aitrendings/core/analyzer"
	"aitrendings/core/cache"
	"aitrendings/core/github"
	"aitrendings/core/model"
	"aitrendings/core/risk"
	"aitrendings/core/scorer"
	"aitrendings/core/store"
)

// PluginExecutor 插件执行器接口
type PluginExecutor interface {
	Execute(ctx context.Context, report *model.Report) error
}

// Pipeline 核心流水线
type Pipeline struct {
	github       github.Fetcher
	analyzer     analyzer.Analyzer
	executor     PluginExecutor
	store        *store.Store
	scorer       *scorer.AdoptionScorer
	riskAssessor *risk.Assessor
	cache        *cache.ReadmeCache
	concurrency  int
}

// Scheduler 定时调度器
type Scheduler struct {
	spec string
	fn   func(ctx context.Context) error
}
