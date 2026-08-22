// Package wechat 提供微信公众号推送插件。
package wechat

import (
	"sync"
	"time"

	pkghttp "aitrendings/pkg/http"
)

// Config 微信公众号配置
type Config struct {
	AppID        string
	Secret       string
	Author       string
	ThumbMediaID string
}

// Plugin 微信公众号插件
type Plugin struct {
	config Config
	client *pkghttp.Client
	mu     sync.Mutex
	token  string
	exp    time.Time
}

// tokenResponse access_token API 响应
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// mediaResponse 素材 API 响应
type mediaResponse struct {
	MediaID string `json:"media_id"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
