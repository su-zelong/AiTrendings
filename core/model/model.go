// Package model 提供核心数据结构定义。
package model

import "time"

// Repo 仓库基本信息
type Repo struct {
	FullName    string
	Name        string
	Owner       string
	Description string
	Stars       int
	Forks       int
	Language    string
	URL         string
}

// Report 分析报告（LLM 输出）
type Report struct {
	RepoName      string         `json:"repo_name"`
	Starts        int            `json:"start"`
	Category      string         `json:"category"`
	Summary       string         `json:"summary"`
	TechStack     []string       `json:"tech_stack"`
	LearningPath  []string       `json:"learning_path"`
	Background    string         `json:"background"`
	RiskFactors   []string       `json:"risk_factors"`
	AdoptionScore *AdoptionScore `json:"adoption_score,omitempty"`
	RiskLevel     string         `json:"risk_level,omitempty"`
	RiskReason    string         `json:"risk_reason,omitempty"`
}

// AdoptionScore 开发者采纳指数
type AdoptionScore struct {
	Total     int            `json:"total"`
	Breakdown map[string]int `json:"breakdown"`
}

// Snapshot 每日快照（存储到数据库）
type Snapshot struct {
	ProjectName  string         `json:"project_name"`
	Category     string         `json:"category"`
	Stars        int            `json:"stars"`
	Forks        int            `json:"forks"`
	DeltaStars   int            `json:"deltaStars"`
	SnapshotDate time.Time      `json:"snapshot_date"`
	RawMetadata  map[string]any `json:"raw_metadata,omitempty"`
}

// CachedAnalysis LLM 缓存
type CachedAnalysis struct {
	ReadmeMD5   string       `json:"readme_md5"`
	LLMResponse *LLMResponse `json:"llm_response"`
	CreatedAt   time.Time    `json:"created_at"`
	ExpiresAt   time.Time    `json:"expires_at"`
}

// LLMResponse LLM 完整响应
type LLMResponse struct {
	Category     string   `json:"category"`
	Summary      string   `json:"summary"`
	TechStack    []string `json:"tech_stack"`
	LearningPath []string `json:"learning_path"`
	Background   string   `json:"background"`
	RiskFactors  []string `json:"risk_factors"`
}

// CategoryGroup 按类别分组的项目列表
type CategoryGroup struct {
	Name  string      `json:"name"`
	Items []*Snapshot `json:"items"`
}

// ComparisonReport 横向对比报告
type ComparisonReport struct {
	Date   time.Time        `json:"date"`
	Groups []*CategoryGroup `json:"groups"`
}

// PushMessage 推送消息
type PushMessage struct {
	ProjectName string
	Stars       int
	Content     string
	URL         string
	PushedAt    time.Time
}
