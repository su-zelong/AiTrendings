// Package github 提供 GitHub API 相关功能。
package github

import (
	"context"
	"net/http"

	"aitrendings/core/model"
)

// Fetcher 负责获取 GitHub 数据。
type Fetcher interface {
	FetchTrending(ctx context.Context) ([]model.Repo, error)
	FetchReadme(ctx context.Context, repo model.Repo) (string, error)
}

// Client GitHub API 客户端
type Client struct {
	token       string
	topN        int
	createdDays int
	client      *http.Client
}

// searchResponse GitHub Search API 响应
type searchResponse struct {
	Items []searchItem `json:"items"`
}

// searchItem GitHub Search API 单项结果
type searchItem struct {
	FullName    string `json:"full_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Stargazers  int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Language    string `json:"language"`
	HTMLURL     string `json:"html_url"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}
