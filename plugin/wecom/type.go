// Package wecom 提供企业微信推送插件。
package wecom

import (
	pkghttp "aitrendings/pkg/http"
)

// Config 企业微信配置
type Config struct {
	Webhook string
}

// Plugin 企业微信插件
type Plugin struct {
	config Config
	client *pkghttp.Client
}

// wecomResponse 企业微信 API 响应
type wecomResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
