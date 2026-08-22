// Package main 提供微信公众号缩略图上传工具。
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	maxThumbSize = 64 * 1024 // 64KB
	apiBase      = "https://api.weixin.qq.com"
)

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

func main() {
	imgPath := flag.String("img", "docs/Flowers.jpg", "缩略图图片路径（JPG格式，64KB以内）")
	force := flag.Bool("force", false, "强制重新上传（即使已配置）")
	flag.Parse()

	// 加载 .env
	envPath := ".env"
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("错误：加载 .env 失败: %v\n", err)
		os.Exit(1)
	}

	appID := os.Getenv("WECHAT_APPID")
	secret := os.Getenv("WECHAT_SECRET")
	if appID == "" || secret == "" {
		fmt.Println("错误：请在 .env 中配置 WECHAT_APPID 和 WECHAT_SECRET")
		os.Exit(1)
	}

	// 检查是否已配置
	if existing := os.Getenv("WECHAT_THUMB_MEDIA_ID"); existing != "" && !*force {
		fmt.Printf("已配置 WECHAT_THUMB_MEDIA_ID: %s\n", existing)
		fmt.Println("如需重新上传，请使用 -force 参数")
		return
	}

	// 读取并压缩图片
	fmt.Printf("读取图片: %s\n", *imgPath)
	imgData, err := compressImage(*imgPath)
	if err != nil {
		fmt.Printf("错误：处理图片失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("压缩后大小: %.2f KB\n", float64(len(imgData))/1024)

	// 获取 access_token
	fmt.Println("获取 access_token...")
	token, err := getAccessToken(appID, secret)
	if err != nil {
		fmt.Printf("错误：获取 access_token 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("access_token 获取成功")

	// 上传缩略图
	fmt.Println("上传缩略图素材...")
	mediaID, err := uploadThumb(token, imgData)
	if err != nil {
		fmt.Printf("错误：上传缩略图失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("上传成功！media_id: %s\n", mediaID)

	// 更新 .env
	if err := updateEnv("WECHAT_THUMB_MEDIA_ID", mediaID); err != nil {
		fmt.Printf("错误：更新 .env 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("已更新 .env 文件")
}

// compressImage 读取图片并压缩到 64KB 以内
func compressImage(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 解码图片
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	// 缩小到合适尺寸（900x500 是公众号封面推荐尺寸）
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w > 900 || h > 500 {
		ratio := min(float64(900)/float64(w), float64(500)/float64(h))
		newW := int(float64(w) * ratio)
		newH := int(float64(h) * ratio)
		resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				srcX := int(float64(x) / ratio)
				srcY := int(float64(y) / ratio)
				resized.Set(x, y, img.At(srcX, srcY))
			}
		}
		img = resized
	}

	// 逐步降低质量直到小于 64KB
	for quality := 95; quality >= 5; quality -= 5 {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("编码 JPEG 失败: %w", err)
		}
		if buf.Len() <= maxThumbSize {
			return buf.Bytes(), nil
		}
	}

	return nil, fmt.Errorf("无法压缩到 %dKB 以内", maxThumbSize/1024)
}

// getAccessToken 获取微信 access_token
func getAccessToken(appID, secret string) (string, error) {
	endpoint := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		apiBase, url.QueryEscape(appID), url.QueryEscape(secret))

	resp, err := http.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result tokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信返回错误: %d %s", result.ErrCode, result.ErrMsg)
	}
	return result.AccessToken, nil
}

// uploadThumb 上传缩略图素材
func uploadThumb(token string, imgData []byte) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("media", "thumb.jpg")
	if err != nil {
		return "", fmt.Errorf("创建表单文件失败: %w", err)
	}
	if _, err := fw.Write(imgData); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	if err := mw.Close(); err != nil {
		return "", fmt.Errorf("关闭 multipart 失败: %w", err)
	}

	endpoint := fmt.Sprintf("%s/cgi-bin/material/add_material?access_token=%s&type=thumb", apiBase, token)
	req, err := http.NewRequest("POST", endpoint, &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result mediaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信返回错误: %d %s", result.ErrCode, result.ErrMsg)
	}
	return result.MediaID, nil
}

// updateEnv 更新 .env 文件中的指定变量
func updateEnv(key, value string) error {
	envPath := ".env"
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("读取 .env 失败: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}
	if !found {
		// 确保最后一行有换行再追加
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, key+"="+value)
	}

	result := strings.Join(lines, "\n")
	// 确保文件以换行结尾
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return os.WriteFile(envPath, []byte(result), 0644)
}
