// Package main 是 AiTreadings 的主入口，负责启动 Agent 并执行定时任务。
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"aitrendings/config"
	"aitrendings/core/analyzer"
	"aitrendings/core/github"
	"aitrendings/core/pipeline"
	"aitrendings/core/store"
	"aitrendings/pkg/db"
	"aitrendings/plugin"

	// 导入插件（触发 init 注册）
	_ "aitrendings/plugin/wecom"
	_ "aitrendings/plugin/wechat"
)

func main() {
	once := flag.Bool("once", false, "run the pipeline once and exit")
	flag.Parse()

	// 1. 加载配置
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.New(ctx, db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		slog.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	// 2. 数据库连接池封装
	bizStore := store.New(database)

	initPlugins(cfg)

	executor := plugin.NewExecutor(plugin.DefaultRegistry)

	g := github.NewClient(cfg.GithubToken, cfg.AnalyzeTopN)
	a := analyzer.NewLLMAnalyzer(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)

	// pipeline
	pl := pipeline.New(g, a, executor, bizStore)

	// 单次执行测试
	if *once {
		if err := pl.Run(ctx); err != nil {
			slog.Error("pipeline run failed", "err", err)
			os.Exit(1)
		}
		return
	}

	// 非单次执行--走定时执行
	sched := pipeline.NewScheduler(cfg.CronSpec, pl.Run)
	slog.Info("agent started", "cron", cfg.CronSpec, "store", bizStore != nil)
	sched.Start(ctx)
}

func initPlugins(cfg *config.Config) {
	pluginConfigs := []plugin.Config{
		{
			Name:    "wecom",
			Enabled: cfg.PushChannel == "wecom" || cfg.PushChannel == "all",
			Options: map[string]interface{}{
				"webhook": cfg.WecomWebhook,
			},
		},
		{
			Name:    "wechat",
			Enabled: cfg.PushChannel == "wechat" || cfg.PushChannel == "all",
			Options: map[string]interface{}{
				"appid":          cfg.WechatAppID,
				"secret":         cfg.WechatSecret,
				"author":         cfg.WechatAuthor,
				"thumb_media_id": cfg.WechatThumbMediaID,
			},
		},
	}

	if err := plugin.DefaultRegistry.Init(pluginConfigs); err != nil {
		slog.Error("init plugins failed", "err", err)
	}
}
