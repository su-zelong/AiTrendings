// Package risk 提供项目风险评估功能。
package risk

import (
	"time"

	"aitrendings/core/model"
)

// New 创建 Assessor
func New() *Assessor {
	return &Assessor{}
}

// Assess 评估项目风险
func (a *Assessor) Assess(repo model.Repo, meta map[string]any) (string, string) {
	var reasons []string

	// 许可证检查
	license := getString(meta, "license")
	if license == "" || license == "NOASSERTION" || license == "Other" {
		reasons = append(reasons, "许可证未明确")
	}

	// Release 检查：近6个月无正式 release
	lastRelease := getTime(meta, "last_release")
	if lastRelease != nil && !lastRelease.IsZero() {
		if daysSince(*lastRelease) > 180 {
			reasons = append(reasons, "近6个月无正式release")
		}
	}

	// 维护压力：open_issues > stars * 10%
	openIssues := getInt(meta, "open_issues")
	if repo.Stars > 0 && float64(openIssues) > float64(repo.Stars)*0.1 {
		reasons = append(reasons, "open_issues > 10% stars")
	}

	// 确定风险等级
	level := "低"
	if len(reasons) >= 2 {
		level = "高"
	} else if len(reasons) == 1 {
		level = "中"
	}

	reason := ""
	if len(reasons) > 0 {
		reason = reasons[0]
		for _, r := range reasons[1:] {
			reason += "；" + r
		}
	}

	return level, reason
}

// getString 从 map 获取 string 值
func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
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
func getTime(m map[string]any, key string) *time.Time {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		return &val
	case string:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
}

// daysSince 计算距今天数
func daysSince(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}
