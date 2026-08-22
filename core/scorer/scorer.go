// Package scorer 提供开发者采纳指数计算功能。
package scorer

import (
	"context"
	"math"
	"time"

	"aitrendings/core/model"
)

// New 创建 AdoptionScorer
func New(store Store) *AdoptionScorer {
	return &AdoptionScorer{store: store}
}

// Score 计算采纳指数
func (s *AdoptionScorer) Score(ctx context.Context, repo model.Repo, meta map[string]any) *model.AdoptionScore {
	breakdown := map[string]int{}

	// 基础因子 (60分)
	breakdown["stars"] = normalizeStars(repo.Stars)
	breakdown["forks"] = normalizeForks(repo.Forks)
	breakdown["issues"] = normalizeIssues(getInt(meta, "open_issues"))
	breakdown["contributors"] = normalizeContributors(getInt(meta, "contributors"))

	total := breakdown["stars"] + breakdown["forks"] + breakdown["issues"] + breakdown["contributors"]

	// 热力加速度 (30分) - 需要至少3天数据
	starDelta := s.store.GetStarDelta(ctx, repo.FullName, 3)
	if starDelta > 0 && repo.Stars > starDelta {
		growth := float64(starDelta) / float64(repo.Stars-starDelta)
		heatScore := int(math.Min(growth*100, 30))
		breakdown["heat"] = heatScore
		total += heatScore
	}

	// 活跃度衰减
	lastPush := getTime(meta, "last_push")
	if !lastPush.IsZero() && time.Since(lastPush) > 90*24*time.Hour {
		total = int(float64(total) * 0.7)
		breakdown["decay"] = -30
	}

	return &model.AdoptionScore{
		Total:     clamp(total, 0, 100),
		Breakdown: breakdown,
	}
}

// normalizeStars stars 标准化到 0-25 分
func normalizeStars(stars int) int {
	// 10000 stars 满分
	score := float64(stars) / 10000 * 25
	return int(math.Min(score, 25))
}

// normalizeForks forks 标准化到 0-15 分
func normalizeForks(forks int) int {
	// 1000 forks 满分
	score := float64(forks) / 1000 * 15
	return int(math.Min(score, 15))
}

// normalizeIssues issues 标准化到 0-10 分（越少越好）
func normalizeIssues(issues int) int {
	if issues == 0 {
		return 10
	}
	// 500 issues 以上为 0 分
	score := 10 - float64(issues)/500*10
	return int(math.Max(score, 0))
}

// normalizeContributors contributors 标准化到 0-10 分
func normalizeContributors(contributors int) int {
	// 100 contributors 满分
	score := float64(contributors) / 100 * 10
	return int(math.Min(score, 10))
}

// getInt 从 map 获取 int 值
func getInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	}
	return 0
}

// getTime 从 map 获取时间
func getTime(m map[string]any, key string) time.Time {
	if m == nil {
		return time.Time{}
	}
	v, ok := m[key]
	if !ok {
		return time.Time{}
	}
	switch val := v.(type) {
	case time.Time:
		return val
	case string:
		t, _ := time.Parse(time.RFC3339, val)
		return t
	}
	return time.Time{}
}

// clamp 限制值在范围内
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
