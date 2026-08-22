// Package analyzer 提供 LLM 分析功能。
package analyzer

import (
	"context"

	"aitrendings/core/model"
	pkghttp "aitrendings/pkg/http"
)

// Analyzer 负责基于 README 生成结构化学习报告。
type Analyzer interface {
	Analyze(ctx context.Context, repo model.Repo, readme string) (*model.Report, error)
}

// LLMAnalyzer 调用 OpenAI 兼容的 chat/completions 接口，要求模型输出 JSON 报告。
type LLMAnalyzer struct {
	baseURL  string
	apiKey   string
	model    string
	client   *pkghttp.Client
	chatPath string
}

// llmResponse LLM API 响应结构
type llmResponse struct {
	Choices []llmChoice `json:"choices"`
}

// llmChoice LLM 响应选项
type llmChoice struct {
	Message llmMessage `json:"message"`
}

// llmMessage LLM 响应消息
type llmMessage struct {
	Content string `json:"content"`
}
