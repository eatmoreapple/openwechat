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

func getFileExt(name string) string {
	ext := filepath.Ext(name)
	if len(ext) == 0 {
		ext = "undefined"
	}
	return strings.TrimPrefix(ext, ".")
}

const (
	pic   = "pic"
	video = "video"
	doc   = "doc"
)

// 微信匹配文件类型策略
func getMessageType(filename string) string {
	ext := getFileExt(filename)
	if _, ok := imageType[ext]; ok {
		return pic
	}
	if ext == videoType {
		return video
	}
	return doc
}
