package openwechat

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ToBuffer(v interface{}) (*bytes.Buffer, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(buf), nil
}

// 获取随机设备id
func GetRandomDeviceId() string {
	rand.Seed(time.Now().Unix())
	var builder strings.Builder
	builder.WriteString("e")
	for i := 0; i < 15; i++ {
		r := rand.Intn(9)
		builder.WriteString(strconv.Itoa(r))
	}
	return builder.String()
}

func getWebWxDataTicket(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "webwx_data_ticket" {
			return cookie.Value
		}
	}
	return ""
}

// Form Xml 格式化
func XmlFormString(text string) string {
	lt := strings.ReplaceAll(text, "&lt;", "<")
	gt := strings.ReplaceAll(lt, "&gt;", ">")
	br := strings.ReplaceAll(gt, "<br/>", "\n")
	return strings.ReplaceAll(br, "&amp;amp;", "&")
}

func getTotalDuration(delay ...time.Duration) time.Duration {
	var total time.Duration
	for _, d := range delay {
		total += d
	}
	return total
}

// 获取文件上传的类型
func GetFileContentType(file multipart.File) (string, error) {
	data := make([]byte, 512)
	if _, err := file.Read(data); err != nil {
		return "", err
	}
	return http.DetectContentType(data), nil
}

func getFileExt(name string) string {
	results := strings.Split(name, ".")
	if len(results) == 0 {
		return "undefined"
	}
	return results[len(results)-1]
}
