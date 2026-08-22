// Package reporter 提供报告生成功能。
package reporter

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"aitrendings/core/model"
)

// New 创建 ComparisonReporter
func New(store Store) *ComparisonReporter {
	return &ComparisonReporter{store: store}
}

// Generate 生成横向对比报告
func (r *ComparisonReporter) Generate(ctx context.Context) (*model.ComparisonReport, error) {
	snapshots, err := r.store.GetRecentSnapshots(ctx, 7)
	if err != nil {
		return nil, err
	}

	// 按 category 分组
	byCategory := map[string][]*model.Snapshot{}
	for _, s := range snapshots {
		cat := s.Category
		if cat == "" {
			cat = "Other"
		}
		byCategory[cat] = append(byCategory[cat], s)
	}

	// 每组按 adoption_score 排序
	report := &model.ComparisonReport{
		Date: time.Now(),
	}

	for category, items := range byCategory {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Stars > items[j].Stars
		})

		topN := items
		if len(topN) > 5 {
			topN = topN[:5]
		}

		report.Groups = append(report.Groups, &model.CategoryGroup{
			Name:  category,
			Items: topN,
		})
	}

	// 按项目数排序
	sort.Slice(report.Groups, func(i, j int) bool {
		return len(report.Groups[i].Items) > len(report.Groups[j].Items)
	})

	return report, nil
}

// RenderMarkdown 渲染 Markdown 格式的对比表格
func (r *ComparisonReporter) RenderMarkdown(report *model.ComparisonReport) string {
	var sb strings.Builder

	sb.WriteString("## 横向对比（按类别）\n\n")
	sb.WriteString(fmt.Sprintf("*数据截至: %s*\n\n", report.Date.Format("2006-01-02")))

	for _, group := range report.Groups {
		sb.WriteString(fmt.Sprintf("### %s\n\n", group.Name))
		sb.WriteString("| 项目 | Stars | Forks | 说明 |\n")
		sb.WriteString("|------|-------|-------|------|\n")

		for _, item := range group.Items {
			desc := ""
			if item.RawMetadata != nil {
				if d, ok := item.RawMetadata["description"].(string); ok {
					desc = truncate(d, 50)
				}
			}
			sb.WriteString(fmt.Sprintf("| %s | %d | %d | %s |\n",
				item.ProjectName, item.Stars, item.Forks, desc))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
