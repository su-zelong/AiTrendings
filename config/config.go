// Package config 提供配置加载功能。
package config

import (
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	LLMBaseURL         string
	LLMAPIKey          string
	LLMModel           string
	GithubToken        string
	WecomWebhook       string
	WechatAppID        string
	WechatSecret       string
	WechatAuthor       string
	WechatThumbMediaID string
	PushChannel        string
	AnalyzeTopN        int
	CronSpec           string

	// 数据库配置
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		LLMBaseURL:         env("LLM_BASE_URL", ""),
		LLMAPIKey:          env("LLM_API_KEY", ""),
		LLMModel:           env("LLM_MODEL", ""),
		GithubToken:        env("GITHUB_TOKEN", ""),
		WecomWebhook:       env("WECOM_WEBHOOK", ""),
		WechatAppID:        env("WECHAT_APPID", ""),
		WechatSecret:       env("WECHAT_SECRET", ""),
		WechatAuthor:       env("WECHAT_AUTHOR", ""),
		WechatThumbMediaID: env("WECHAT_THUMB_MEDIA_ID", ""),
		PushChannel:        env("PUSH_CHANNEL", ""),
		AnalyzeTopN:        envInt("ANALYZE_TOP_N", 10),
		CronSpec:           env("CRON_SPEC", "0 8 * * *"),

		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     envInt("DB_PORT", 5432),
		DBUser:     env("DB_USER", "postgres"),
		DBPassword: env("DB_PASSWORD", ""),
		DBName:     env("DB_NAME", "aitrendings"),
		DBSSLMode:  env("DB_SSLMODE", "disable"),
	}
}

// 配置项从环境变量导入
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// 环境变量导入int类型--port等
func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return def
}
