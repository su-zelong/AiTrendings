// Package pipeline 提供核心流水线编排功能。
package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"aitrendings/core/analyzer"
	"aitrendings/core/cache"
	"aitrendings/core/github"
	"aitrendings/core/model"
	"aitrendings/core/risk"
	"aitrendings/core/scorer"
	"aitrendings/core/store"
)

// defaultConcurrency 控制同时分析的仓库数
const defaultConcurrency = 3

// New 创建流水线
func New(
	g github.Fetcher,
	a analyzer.Analyzer,
	executor PluginExecutor,
	s *store.Store,
) *Pipeline {
	return &Pipeline{
		github:       g,
		analyzer:     a,
		executor:     executor,
		store:        s,
		scorer:       scorer.New(s),
		riskAssessor: risk.New(),
		cache:        cache.New(s),
		concurrency:  defaultConcurrency,
	}
}

// Run 执行一次流水线
func (p *Pipeline) Run(ctx context.Context) error {
	// 1. 获取所有仓库
	repos, err := p.github.FetchTrending(ctx)
	if err != nil {
		return err
	}
	slog.Info("fetched trending repos", "count", len(repos))

	var (
		mu      sync.Mutex
		reports []*model.Report
		fails   int
	)

	// 2. 一次分析仓库数量切片，default=3
	sem := make(chan struct{}, p.concurrency)
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(repo model.Repo) {
			defer wg.Done()
			defer func() { <-sem }()
			// 2.1 分析报告；每次起三个进程分析
			report, err := p.analyzeRepo(ctx, repo)
			if err != nil {
				mu.Lock()
				fails++
				mu.Unlock()
				slog.Error("analyze repo failed", "repo", repo.FullName, "err", err)
				return
			}

			mu.Lock()
			reports = append(reports, report)
			mu.Unlock()
			slog.Info("repo analyzed", "repo", repo.FullName)
		}(repo)
	}
	wg.Wait()

	if len(reports) == 0 {
		return fmt.Errorf("no reports generated")
	}

	// 保存快照并生成报告
	for _, report := range reports {
		// 计算 Adoption Score
		repo := model.Repo{
			FullName: report.RepoName,
			Stars:    report.Starts,
		}
		score := p.scorer.Score(ctx, repo, nil)
		report.AdoptionScore = score

		// 风险评估
		riskLevel, riskReason := p.riskAssessor.Assess(repo, nil)
		report.RiskLevel = riskLevel
		report.RiskReason = riskReason

		// 保存快照
		snap := &model.Snapshot{
			ProjectName:  report.RepoName,
			Category:     report.Category,
			SnapshotDate: time.Now(),
			RawMetadata: map[string]any{
				"summary":       report.Summary,
				"tech_stack":    report.TechStack,
				"learning_path": report.LearningPath,
				"background":    report.Background,
				"risk_factors":  report.RiskFactors,
			},
		}
		if err := p.store.SaveSnapshot(ctx, snap); err != nil {
			slog.Error("save snapshot failed", "repo", report.RepoName, "err", err)
		}
	}

	for _, report := range reports {
		if err := p.executor.Execute(ctx, report); err != nil {
			slog.Error("plugin execution failed", "repo", report.RepoName, "err", err)
		}
	}

	slog.Info("pipeline finished", "total", len(repos), "analyzed", len(reports), "failed", fails)
	return nil
}

func (p *Pipeline) analyzeRepo(ctx context.Context, repo model.Repo) (*model.Report, error) {
	content, err := p.github.FetchReadme(ctx, repo)
	if err != nil {
		return nil, err
	}

	// 检查缓存
	if cached, _ := p.cache.Get(ctx, content); cached != nil {
		slog.Info("using cached analysis", "repo", repo.FullName)
		report := &model.Report{
			RepoName:     repo.FullName,
			Starts:       repo.Stars,
			Category:     cached.Category,
			Summary:      cached.Summary,
			TechStack:    cached.TechStack,
			LearningPath: cached.LearningPath,
			Background:   cached.Background,
			RiskFactors:  cached.RiskFactors,
		}
		return report, nil
	}

	// 调用 LLM 分析
	report, err := p.analyzer.Analyze(ctx, repo, content)
	if err != nil {
		return nil, err
	}

	// 保存缓存
	llmResp := &model.LLMResponse{
		Category:     report.Category,
		Summary:      report.Summary,
		TechStack:    report.TechStack,
		LearningPath: report.LearningPath,
		Background:   report.Background,
		RiskFactors:  report.RiskFactors,
	}
	if err := p.cache.Set(ctx, content, llmResp); err != nil {
		slog.Error("set cache failed", "repo", repo.FullName, "err", err)
	}

	return report, nil
}

// NewScheduler 创建调度器
func NewScheduler(spec string, fn func(ctx context.Context) error) *Scheduler {
	return &Scheduler{spec: spec, fn: fn}
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	c := cron.New()
	_, err := c.AddFunc(s.spec, func() {
		if err := s.fn(ctx); err != nil {
			slog.Error("scheduled task failed", "err", err)
		}
	})
	if err != nil {
		slog.Error("add cron job failed", "err", err)
		return
	}

	c.Start()
	slog.Info("scheduler started", "spec", s.spec)
	<-ctx.Done()
	c.Stop()
}
