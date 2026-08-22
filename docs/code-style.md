# Go 项目代码规范

本文档是本项目的统一编码规范，所有新增/修改代码必须遵守。规范优先级：本文档 > 官方 `gofmt` > 常识。

---

## 1. 工具链

| 工具 | 用途 | 命令 |
|------|------|------|
| `gofmt` | 格式化 | `gofmt -l .`（检查） / `gofmt -w .`（修复） |
| `go vet` | 静态检查 | `go vet ./...` |
| `go build` | 编译 | `go build ./...` |
| `golangci-lint`（可选） | 综合 lint | `golangci-lint run` |

**提交/完成前必须通过**：`gofmt` 无差异、`go vet` 无报错、`go build` 成功、相关包测试通过。

---

## 2. 项目结构

```
├── cmd/                    # 可执行程序入口，一个目录一个 main 包
├── internal/               # 内部包，不允许外部导入
│   ├── <module>/           # 按业务模块划分（trending、readme、analyzer...）
│   │   ├── xxx.go          # 接口 + 实现
│   │   └── xxx_test.go     # 同包测试
├── docs/                   # 文档
└── go.mod
```

规则：
- `cmd/` 下的 main 只负责**组装依赖、启动**，不写业务逻辑。
- `internal/` 下按模块分包，模块间通过接口解耦，禁止循环导入。
- 包名小写单数，与目录名一致。

---

## 3. 命名规范

| 类别 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写、单数、无下划线 | `trending`、`readme` |
| 类型 | 大驼峰（PascalCase） | `GitHubTrending` |
| 导出函数/方法 | 大驼峰 | `Fetch`、`NewGitHubTrending` |
| 未导出标识符 | 小驼峰（camelCase） | `retryDelay`、`doRequest` |
| 常量 | 小驼峰或全大写均可，同包内统一 | `searchRepoURL` |
| 接口名 | 行为名或 `Xxxer` | `Fetcher`、`Analyzer` |
| 缩写 | 全大写（URL、ID、API、HTTP） | `APIKey`、`LLMModel` |
| 错误变量 | 前缀 `Err` | `ErrNotFound` |

---

## 4. 代码风格

- **必须** 由 `gofmt`/`goimports` 格式化（tab 缩进）。
- 行宽尽量 ≤ 120 字符，过长时换行。
- 导入分组顺序：标准库 → 第三方库 → 本项目包，组间空行。
- 导出标识符必须有文档注释（`// Fetch 获取...`），注释与被注释代码对齐。
- **禁止**无意义注释（如 `// 定义变量`）；注释解释"为什么"，不解释"做了什么"。
- 结构体字段对齐，JSON 标签放右侧：`Stars int \`json:"stars"\``。
- 声明与使用尽量靠近，减少裸类型混用。

---

## 5. 错误处理

- 每个可能失败的分支都必须处理，**禁止吞掉错误**（`_ = fn()` 除非明确必要）。
- 包装错误保留上下文：`fmt.Errorf("fetch trending: %w", err)`。
- 错误信息小写开头，不结尾加标点。
- 自定义哨兵错误用 `errors.New("...")` 并导出（`ErrXxx`），配合 `errors.Is` 判断。
- 可重试操作（网络/限流）要有重试逻辑，重试间隔可配置、默认指数退避。
- 外部系统错误要保留状态码等关键信息，方便排查。

---

## 6. HTTP 客户端

- 所有 HTTP 调用统一设置 `Timeout`，禁止无超时请求。
- 统一使用 `context.Context`，可被上层取消。
- 限流（403/429）需处理并退避重试。
- 读取响应后及时 `Close()`，可用 `defer` 或显式关闭。
- 请求头、URL 等参数尽量收敛到模块内部，不散落业务层。

---

## 7. 并发与数据

- 需要并发时用 goroutine + channel/WaitGroup，控制并发上限（如 `errgroup.SetLimit`）。
- 共享数据用 `sync.Mutex`/`atomic` 保护，优先用 channel 传递所有权。
- 避免在循环内无限制创建 goroutine。
- 所有并发逻辑必须处理 `context` 取消。

---

## 8. 配置管理

- 配置统一从环境变量读取，集中在 `internal/config`，不允许在业务代码中直接 `os.Getenv`。
- 敏感信息（Token、Key）只走环境变量/密钥管理，**禁止硬编码或提交到仓库**。
- 新增配置时同步更新 `.env.example`。
- 配置字段有默认值，缺省时运行不报错（除必需项）。

---

## 9. 测试

- 每个模块至少覆盖：正常路径 + 至少一个错误路径。
- 外部 API 调用用 `httptest.NewServer` mock，**禁止测试依赖真实网络**。
- 测试命名 `Test<函数名>`，子场景用 `t.Run("描述", ...)`。
- 有状态/重试逻辑要可配置参数，保证测试不慢（如重试间隔设小）。
- 测试结束后恢复全局被篡改的变量（如 `defer` 还原 URL）。

---

## 10. 提交规范

- 提交信息格式：`type(scope): description`
- type：`feat` / `fix` / `refactor` / `test` / `docs` / `chore` / `perf`
- scope：模块名，如 `trending`、`pipeline`
- 示例：`feat(trending): fetch top trending repos via search api`

---

## 11. AI Agent 项目的额外约定

- 所有外部系统（GitHub、LLM、Webhook）都封装在独立模块内，业务层只面对接口。
- LLM 相关：prompt 不散落在代码中，统一放 `internal/<module>/prompt.go` 或独立目录。
- 每个模块预留 Agent 扩展点：接口定义清晰、单职责、副作用最小。
- 日志用 `log/slog`，记录关键流程节点，禁止打印敏感信息（Token、Key）。
- 关注成本：LLM 调用前先判断是否必要，避免重复请求。
