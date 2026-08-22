// Package analyzer 提供 LLM 分析功能。
package analyzer

import (
	"fmt"
	"strings"

	"aitrendings/core/model"
)

// buildSystemPrompt 返回 LLM 的系统提示词，要求模型以指定 JSON 结构输出。
func buildSystemPrompt() string {
	return `你是一个资深开源项目分析专家。你会收到一个 GitHub 仓库的 README 原文，请基于它生成一份结构化学习报告。

报告必须严格输出为 JSON 对象，包含以下字段：

- category: 项目分类，从以下选项中选择一个最合适的：
  Web Framework, Database, CLI Tool, AI/ML, DevOps, Library, SDK, 
  Security, Monitoring, Testing, Documentation, Game, Mobile, Other

- summary: 详细项目介绍（150-200字），必须包含以下三个维度：
  1. 项目是什么：一句话定义项目类型（库/框架/工具/平台/应用），核心功能是什么
  2. 技术背景：解决什么问题/痛点，现有方案有什么不足，为什么需要这个项目
  3. 价值与特色：核心优势是什么，适合什么场景，有什么独特之处

- tech_stack: 项目使用的核心技术栈，字符串数组

- learning_path: 建议的学习路线，按从入门到深入的顺序排列的字符串数组（6-8条），必须包含：
  1. 基础知识：需要掌握哪些前置技能
  2. 前置课程：推荐完成哪些文档或教程
  3. 入门步骤：如何快速上手运行项目
  4. 核心文件：建议先阅读哪些关键文件
  5. 关键概念：需要理解的核心技术原理
  6. 深入学习：如何进阶学习

- background: 项目的背景、解决的问题或定位（2-3 句话）

- risk_factors: 项目的潜在风险因素，字符串数组，例如：
  - "许可证不明确"
  - "维护不活跃"
  - "依赖过多"
  - "文档不完善"
  - "社区较小"
  如果没有明显风险，返回空数组 []

只输出 JSON，不要输出任何其他内容。`
}

// buildUserPrompt 构造针对单个仓库的请求内容。
func buildUserPrompt(repo model.Repo, readme string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "仓库: %s\n", repo.FullName)
	if repo.Description != "" {
		fmt.Fprintf(&sb, "描述: %s\n", repo.Description)
	}
	fmt.Fprintf(&sb, "语言: %s\n", repo.Language)
	fmt.Fprintf(&sb, "Star: %d\n", repo.Stars)
	fmt.Fprintf(&sb, "\n===== README 原文 =====\n%s\n", readme)
	return sb.String()
}
