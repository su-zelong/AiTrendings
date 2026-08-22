// Package github 提供 GitHub API 相关功能。
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"aitrendings/core/model"
)

var (
	ErrNotFound = errors.New("readme not found")
	searchURL   = "https://api.github.com/search/repositories"
	apiBase     = "https://api.github.com"
)

// NewClient 创建 GitHub 客户端
func NewClient(token string, topN int) *Client {
	return &Client{
		token:       token,
		topN:        topN,
		createdDays: 7,
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchTrending 获取热门项目
func (c *Client) FetchTrending(ctx context.Context) ([]model.Repo, error) {
	since := time.Now().AddDate(0, 0, -c.createdDays).Format("2006-01-02")

	q := url.Values{}
	q.Set("q", "created:>"+since)
	q.Set("sort", "stars")
	q.Set("order", "desc")
	q.Set("per_page", strconv.Itoa(c.topN))

	endpoint := searchURL + "?" + q.Encode()

	body, err := c.doGet(ctx, endpoint, "application/vnd.github+json")
	if err != nil {
		return nil, err
	}

	var resp searchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	repos := make([]model.Repo, 0, len(resp.Items))
	for _, it := range resp.Items {
		repos = append(repos, model.Repo{
			FullName:    it.FullName,
			Name:        it.Name,
			Owner:       it.Owner.Login,
			Description: it.Description,
			Stars:       it.Stargazers,
			Forks:       it.Forks,
			Language:    it.Language,
			URL:         it.HTMLURL,
		})
	}
	return repos, nil
}

// FetchReadme 获取仓库 README
func (c *Client) FetchReadme(ctx context.Context, repo model.Repo) (string, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/readme", apiBase, repo.Owner, repo.Name)

	body, err := c.doGet(ctx, endpoint, "application/vnd.github.raw")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", fmt.Errorf("%w: %s/%s", ErrNotFound, repo.Owner, repo.Name)
		}
		return "", err
	}
	return string(body), nil
}

// doGet 发起 GET 请求
func (c *Client) doGet(ctx context.Context, endpoint, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "aitrendings-agent")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return body, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	return nil, fmt.Errorf("github api error (status %d): %s", resp.StatusCode, truncate(string(body), 200))
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
