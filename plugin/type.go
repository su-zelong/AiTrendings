// Package plugin 提供插件接口定义和注册机制。
package plugin

import (
	"context"
	"sync"

	"aitrendings/core/model"
)

// Plugin 插件接口
type Plugin interface {
	// Name 插件名称
	Name() string

	// Version 版本号
	Version() string

	// Init 初始化插件
	Init(config Config) error

	// Execute 执行插件
	Execute(ctx context.Context, report *model.Report) error

	// Shutdown 关闭插件
	Shutdown() error
}

// Config 插件配置
type Config struct {
	Name    string
	Enabled bool
	Options map[string]interface{}
}

// Executor 插件执行器
type Executor struct {
	registry *Registry
}

// Registry 插件注册表
type Registry struct {
	mu        sync.RWMutex
	factories map[string]func() Plugin
	plugins   map[string]Plugin
}
