// Package wecom 提供企业微信推送插件。
package wecom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"aitrendings/core/model"
	pkghttp "aitrendings/pkg/http"
	"aitrendings/plugin"
)

func init() {
	plugin.Register("wecom", NewPlugin)
}

// NewPlugin 创建企业微信插件
func NewPlugin() plugin.Plugin {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return &Plugin{
		client: pkghttp.NewClient(httpClient, pkghttp.DefaultConfig),
	}
}

// Name 插件名称
func (p *Plugin) Name() string {
	return "wecom"
}

// Version 版本号
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Init 初始化插件
func (p *Plugin) Init(config plugin.Config) error {
	webhook, ok := config.Options["webhook"].(string)
	if !ok || webhook == "" {
		return fmt.Errorf("wecom: missing webhook")
	}

	p.config = Config{Webhook: webhook}
	slog.Info("wecom plugin initialized")
	return nil
}

// Execute 执行推送
func (p *Plugin) Execute(ctx context.Context, report *model.Report) error {
	content := p.render(report)
	return p.send(ctx, content)
}

// Shutdown 关闭插件
func (p *Plugin) Shutdown() error {
	return nil
}

// render 渲染消息
func (p *Plugin) render(report *model.Report) string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("## 📦 %s\n\n", report.RepoName))

	if report.Summary != "" {
		b.WriteString(fmt.Sprintf("**摘要**：%s\n\n", report.Summary))
	}

	if len(report.TechStack) > 0 {
		b.WriteString("**技术栈**：")
		for i, tech := range report.TechStack {
			if i > 0 {
				b.WriteString("、")
			}
			b.WriteString(tech)
		}
		b.WriteString("\n\n")
	}

	if len(report.LearningPath) > 0 {
		b.WriteString("**学习路线**：\n")
		for i, step := range report.LearningPath {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// send 发送消息
func (p *Plugin) send(ctx context.Context, content string) error {
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	respBody, err := p.client.Do(ctx, http.MethodPost, p.config.Webhook, bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}

	var result wecomResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wecom api error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}
