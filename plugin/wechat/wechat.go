// Package wechat 提供微信公众号推送插件。
package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"aitrendings/core/model"
	pkghttp "aitrendings/pkg/http"
	"aitrendings/plugin"
)

func init() {
	plugin.Register("wechat", NewPlugin)
}

// NewPlugin 创建微信公众号插件
func NewPlugin() plugin.Plugin {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return &Plugin{
		client: pkghttp.NewClient(httpClient, pkghttp.DefaultConfig),
	}
}

// Name 插件名称
func (p *Plugin) Name() string {
	return "wechat"
}

// Version 版本号
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Init 初始化插件
func (p *Plugin) Init(config plugin.Config) error {
	appID, _ := config.Options["appid"].(string)
	secret, _ := config.Options["secret"].(string)
	author, _ := config.Options["author"].(string)
	thumbMediaID, _ := config.Options["thumb_media_id"].(string)

	if appID == "" || secret == "" {
		return fmt.Errorf("wechat: missing appid or secret")
	}

	p.config = Config{
		AppID:        appID,
		Secret:       secret,
		Author:       author,
		ThumbMediaID: thumbMediaID,
	}

	slog.Info("wechat plugin initialized")
	return nil
}

// Execute 执行推送
func (p *Plugin) Execute(ctx context.Context, report *model.Report) error {
	// 获取 access_token
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access_token: %w", err)
	}

	// 获取封面素材 ID
	thumbMediaID := p.config.ThumbMediaID
	if thumbMediaID == "" {
		// 尝试上传默认封面
		thumbMediaID, err = p.uploadDefaultThumb(ctx, token)
		if err != nil {
			slog.Warn("upload thumb failed, using empty", "err", err)
			thumbMediaID = ""
		}
	}

	// 创建草稿
	draftID, err := p.createDraft(ctx, token, report, thumbMediaID)
	if err != nil {
		return fmt.Errorf("create draft: %w", err)
	}

	slog.Info("wechat draft created", "media_id", draftID, "repo", report.RepoName)
	return nil
}

// Shutdown 关闭插件
func (p *Plugin) Shutdown() error {
	return nil
}

// getAccessToken 获取 access_token
func (p *Plugin) getAccessToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.token != "" && time.Now().Before(p.exp) {
		return p.token, nil
	}

	endpoint := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(p.config.AppID), url.QueryEscape(p.config.Secret))

	body, err := p.client.Do(ctx, http.MethodGet, endpoint, nil, "")
	if err != nil {
		return "", err
	}

	var resp tokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("wechat api error: %d %s", resp.ErrCode, resp.ErrMsg)
	}

	p.token = resp.AccessToken
	p.exp = time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - 5*time.Minute)
	return p.token, nil
}

// uploadDefaultThumb 上传默认封面
func (p *Plugin) uploadDefaultThumb(ctx context.Context, token string) (string, error) {
	// 生成一个简单的封面图
	cover, err := p.generateCover()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("media", "cover.jpg")
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(cover); err != nil {
		return "", err
	}
	if err := mw.Close(); err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=%s&type=thumb", token)
	body, err := p.client.Do(ctx, http.MethodPost, endpoint, &buf, mw.FormDataContentType())
	if err != nil {
		return "", err
	}

	var resp mediaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("upload thumb error: %d %s", resp.ErrCode, resp.ErrMsg)
	}

	return resp.MediaID, nil
}

// generateCover 生成简单封面
func (p *Plugin) generateCover() ([]byte, error) {
	return os.ReadFile("docs/Flowers.jpg")
}

// createDraft 创建草稿
func (p *Plugin) createDraft(ctx context.Context, token string, report *model.Report, thumbMediaID string) (string, error) {
	content := p.renderHTML(report)

	article := map[string]any{
		"article_type":   "news",
		"title":          truncate(p.truncateTitle(report.RepoName), 32),
		"content":        content,
		"thumb_media_id": thumbMediaID,
	}
	if p.config.Author != "" {
		article["author"] = truncate(p.config.Author, 16)
	}

	payload := map[string]any{
		"articles": []map[string]any{article},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/draft/add?access_token=%s", token)
	respBody, err := p.client.Do(ctx, http.MethodPost, endpoint, bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}

	var resp mediaResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", err
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("create draft error: %d %s", resp.ErrCode, resp.ErrMsg)
	}

	return resp.MediaID, nil
}

// renderHTML 渲染 HTML 内容
func (p *Plugin) renderHTML(report *model.Report) string {
	var b bytes.Buffer

	b.WriteString("<section style=\"font-size: 15px; color: #333; line-height: 1.8;\">")

	b.WriteString(fmt.Sprintf("<h2 style=\"font-size: 18px; font-weight: 700; color: #333; margin: 0 0 12px 0;\">%s</h2>", report.RepoName))

	if report.Summary != "" {
		b.WriteString(fmt.Sprintf("<section style=\"margin-bottom: 16px;\"><strong style=\"color: #667eea;\">项目简介</strong><p style=\"margin: 6px 0 0 0; color: #495057;\">%s</p></section>", report.Summary))
	}

	if len(report.TechStack) > 0 {
		b.WriteString("<section style=\"margin-bottom: 16px;\"><strong style=\"color: #667eea;\">技术栈</strong><p style=\"margin: 6px 0 0 0; color: #495057;\">")
		for i, tech := range report.TechStack {
			if i > 0 {
				b.WriteString("、")
			}
			b.WriteString(tech)
		}
		b.WriteString("</p></section>")
	}

	if len(report.LearningPath) > 0 {
		b.WriteString("<section style=\"margin-bottom: 16px;\"><strong style=\"color: #667eea;\">学习路线</strong><section style=\"margin: 6px 0 0 0; color: #495057;\">")
		for _, step := range report.LearningPath {
			b.WriteString(fmt.Sprintf("<p style=\"margin: 4px 0;\">• %s</p>", step))
		}
		b.WriteString("</section></section>")
	}

	b.WriteString("</section>")
	return b.String()
}

// truncateTitle 截断标题
func (p *Plugin) truncateTitle(repoName string) string {
	return fmt.Sprintf("GitHub Trending | %s", repoName)
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
