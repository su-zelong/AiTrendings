// Package plugin 提供插件接口定义和注册机制。
package plugin

import (
	"context"
	"log/slog"

	"aitrendings/core/model"
)

// NewExecutor 创建插件执行器
func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

// Execute 执行所有已启用的插件
func (e *Executor) Execute(ctx context.Context, report *model.Report) error {
	plugins := e.registry.GetAll()

	var lastErr error
	for _, p := range plugins {
		if err := p.Execute(ctx, report); err != nil {
			slog.Error("plugin execution failed",
				"plugin", p.Name(),
				"repo", report.RepoName,
				"err", err)
			lastErr = err
			continue
		}
		slog.Info("plugin executed successfully",
			"plugin", p.Name(),
			"repo", report.RepoName)
	}

	return lastErr
}
