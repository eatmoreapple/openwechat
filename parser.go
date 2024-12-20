package openwechat

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func jsonEncode(v interface{}) (io.Reader, error) {
	var buffer = bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	// 这里要设置禁止html转义
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buffer, nil
}

// GetRandomDeviceId 获取随机设备id
func GetRandomDeviceId() string {
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	var builder strings.Builder
	builder.Grow(16)
	builder.WriteString("e")
	for i := 0; i < 15; i++ {
		r := rng.Intn(9)
		builder.WriteString(strconv.Itoa(r))
	}
	return builder.String()
}

// GetFileContentType 获取文件上传的类型
func GetFileContentType(file io.Reader) (string, error) {
	data := make([]byte, 512)
	if _, err := file.Read(data); err != nil {
		return "", err
	}
	return http.DetectContentType(data), nil
}

// fileExtension
func fileExtension(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if len(ext) == 0 {
		return "undefined"
	}
	return strings.TrimPrefix(ext, ".")
}

// 判断是否是图片
func isImageType(imageType string) bool {
	switch imageType {
	case "bmp", "png", "jpeg", "jpg", "gif":
		return true
	default:
		return false
	}
}

// 微信匹配文件类型策略
func messageType(filename string) string {
	ext := fileExtension(filename)
	if isImageType(ext) {
		return "pic"
	}
	if ext == "mp4" {
		return "video"
	}
	return "doc"
}
