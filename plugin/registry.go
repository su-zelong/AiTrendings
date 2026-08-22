// Package plugin 提供插件接口定义和注册机制。
package plugin

import (
	"fmt"
	"log/slog"
)

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]func() Plugin),
		plugins:   make(map[string]Plugin),
	}
}

// DefaultRegistry 默认注册表
var DefaultRegistry = NewRegistry()

// Register 注册插件工厂函数
func Register(name string, factory func() Plugin) {
	DefaultRegistry.mu.Lock()
	defer DefaultRegistry.mu.Unlock()
	DefaultRegistry.factories[name] = factory
	slog.Info("plugin registered", "name", name)
}

// Init 初始化所有插件
func (r *Registry) Init(configs []Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, cfg := range configs {
		factory, ok := r.factories[cfg.Name]
		if !ok {
			slog.Warn("unknown plugin, skipping", "name", cfg.Name)
			continue
		}

		if !cfg.Enabled {
			slog.Info("plugin disabled", "name", cfg.Name)
			continue
		}

		p := factory()
		if err := p.Init(cfg); err != nil {
			return fmt.Errorf("init plugin %s: %w", cfg.Name, err)
		}

		r.plugins[cfg.Name] = p
		slog.Info("plugin initialized", "name", cfg.Name, "version", p.Version())
	}

	return nil
}

// Get 获取插件
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	return p, ok
}

// GetAll 获取所有已初始化的插件
func (r *Registry) GetAll() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var plugins []Plugin
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// Shutdown 关闭所有插件
func (r *Registry) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, p := range r.plugins {
		if err := p.Shutdown(); err != nil {
			slog.Error("shutdown plugin failed", "name", name, "err", err)
		}
	}
}
