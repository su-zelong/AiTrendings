// Package analyzer 提供 LLM 分析功能。
package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aitrendings/core/model"
	pkghttp "aitrendings/pkg/http"
)

// ErrEmptyResponse 表示模型返回了空内容。
var ErrEmptyResponse = errors.New("empty llm response")

// ErrInvalidJSON 表示模型输出无法解析为 Report。
var ErrInvalidJSON = errors.New("invalid llm json output")

// NewLLMAnalyzer 创建分析器，baseURL 形如 https://api.deepseek.com。
func NewLLMAnalyzer(baseURL, apiKey, model string) *LLMAnalyzer {
	httpClient := &http.Client{Timeout: 120 * time.Second}
	return &LLMAnalyzer{
		baseURL:  strings.TrimRight(baseURL, "/"),
		apiKey:   apiKey,
		model:    model,
		client:   pkghttp.NewClient(httpClient, pkghttp.DefaultConfig),
		chatPath: "/chat/completions",
	}
}

// Analyze 解析 README 并生成结构化学习报告。
func (a *LLMAnalyzer) Analyze(ctx context.Context, repo model.Repo, readme string) (*model.Report, error) {
	payload := map[string]any{
		"model": a.model,
		"messages": []map[string]string{
			{"role": "system", "content": buildSystemPrompt()},
			{"role": "user", "content": buildUserPrompt(repo, readme)},
		},
		"temperature":     0.3,
		"response_format": map[string]string{"type": "json_object"},
	}

	body, err := a.doRequest(ctx, payload)
	if err != nil {
		return nil, err
	}

	var resp llmResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse llm response: %w", err)
	}
	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return nil, ErrEmptyResponse
	}

	var report model.Report
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &report); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	report.RepoName = repo.FullName
	return &report, nil
}

// doRequest 发送 chat/completions 请求。
func (a *LLMAnalyzer) doRequest(ctx context.Context, payload map[string]any) ([]byte, error) {
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := a.baseURL + a.chatPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	return a.client.Do(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody), "application/json")
}
