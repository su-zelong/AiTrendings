# AiTreadings

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/DeepSeek-API-4D6BFE?style=for-the-badge&logo=openai&logoColor=white" alt="DeepSeek">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<p align="center">
  <strong>🔥 GitHub Trending 智能情报 Agent</strong>
</p>

<p align="center">
  每日自动抓取 GitHub 热门项目 · 深度分析 · 生成结构化学习报告 · 推送到企业微信/微信公众号
</p>

---

## ✨ 功能特性

- 📊 **智能分析** - 使用 DeepSeek/OpenAI 深度分析项目 README，生成结构化报告
- 🚀 **多通道推送** - 支持企业微信群机器人、微信公众号草稿箱
- ⏰ **定时执行** - Cron 调度，默认每天 08:00 自动运行
- 💾 **本地备份** - 每次运行自动保存 Markdown 和 HTML 到 output/ 目录
- 🔌 **插件化架构** - 核心逻辑与外部集成解耦，新增推送渠道只需添加插件
- 🎯 **独立工具** - 封面上传工具独立运行，支持自动压缩到 64KB

---

## 📁 项目结构

```
AiTreadings/
├── cmd/
│   ├── agent/              # 🎯 主入口
│   └── upload-thumb/       # 🖼️ 封面上传工具
├── config/                 # ⚙️ 配置加载
├── core/                   # 🧠 核心业务逻辑（无外部依赖）
│   ├── analyzer/           # 🧠 LLM 分析器
│   ├── model/              # 📦 数据模型
│   ├── pipeline/           # 🔗 流程编排
│   ├── readme/             # 📄 README 获取
│   ├── scheduler/          # ⏱️ 定时调度
│   └── trending/           # 📈 Trending 获取
├── plugin/                 # 🔌 插件系统（可插拔）
│   ├── interface.go        # 插件接口
│   ├── registry.go         # 插件注册表
│   ├── executor.go         # 插件执行器
│   ├── wechat/             # 微信公众号插件
│   └── wecom/              # 企业微信插件
├── templates/              # 📝 文章模板
└── output/                 # 💾 本地输出
```

---

## 🚀 快速开始

### 环境要求

- Go 1.21+
- DeepSeek 或 OpenAI API Key

### 配置

```bash
cp .env.example .env
```

**必填配置：**

| 变量 | 说明 | 示例 |
|------|------|------|
| `LLM_API_KEY` | API Key | `sk-xxx` |
| `PUSH_CHANNEL` | 推送通道 | `wecom` / `wechat` |

**可选配置：**

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ANALYZE_TOP_N` | 分析项目数 | `10` |
| `CRON_SPEC` | 调度表达式 | `0 8 * * *` |
| `WECHAT_APPID` | 微信公众号 ID | - |
| `WECOM_WEBHOOK` | 企业微信 Webhook | - |

### 运行

```bash
# 编译
go build ./...

# 定时任务模式
go run ./cmd/agent

# 单次执行
go run ./cmd/agent -once
```

---

## 📸 推送效果

### 企业微信

推送内容包含：项目名称、Stars 数量、技术栈标签、学习路线

### 微信公众号草稿

推送内容包含：
- 📝 项目简介（150-200字详细说明）
- 🛠️ 技术栈标签
- 📚 学习路线（6-8条具体指导）
- 🔗 原文链接

本地预览文件：`output/YYYY-MM-DD.html`

---

## 🛠️ 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| LLM | DeepSeek / OpenAI |
| API | GitHub REST API |
| 推送 | 企业微信 / 微信公众号 |
| 调度 | Cron |

---

## 🗺️ 开发路线

```
阶段 0 ──→ 阶段 1 ──→ 阶段 2 ──→ 阶段 3 ──→ 阶段 4
  │          │          │          │          │
  ▼          ▼          ▼          ▼          ▼
硬编码     单 Agent   多 Agent   记忆系统   双向交互
流水线     ReAct      协作分工   向量存储   企业微信
```

| 阶段 | 目标 | 状态 |
|------|------|------|
| 0 | 硬编码流水线骨架 | ✅ 完成 |
| 1 | 单 Agent：ReAct 流程 | 🚀 进行中 |
| 2 | 多 Agent：协作分工 | 📋 计划中 |
| 3 | 记忆：向量数据库 | 📋 计划中 |
| 4 | 交互：双向对话 | 📋 计划中 |

---

## 📚 参考文档

- [企业微信机器人](https://developer.work.weixin.qq.com/document/path/91770)
- [微信公众号草稿接口](https://developers.weixin.qq.com/doc/subscription/api/draftbox/draftmanage/api_draft_add.html)
- [GitHub REST API](https://docs.github.com/rest)
- [DeepSeek API](https://platform.deepseek.com/api-docs)

---

## 📄 License

[MIT](LICENSE)

---

<p align="center">
  <sub>Made with ❤️ by <a href="https://github.com/your-username">AiTreadings</a></sub>
</p>
