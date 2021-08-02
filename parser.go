package openwechat

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"math/rand"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func ToBuffer(v interface{}) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	// 这里要设置禁止html转义
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	return &buffer, err
}

// GetRandomDeviceId 获取随机设备id
func GetRandomDeviceId() string {
	rand.Seed(time.Now().Unix())
	var builder strings.Builder
	builder.Grow(16)
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

func getTotalDuration(delay ...time.Duration) time.Duration {
	var total time.Duration
	for _, d := range delay {
		total += d
	}
	return total
}

// GetFileContentType 获取文件上传的类型
func GetFileContentType(file multipart.File) (string, error) {
	data := make([]byte, 512)
	if _, err := file.Read(data); err != nil {
		return "", err
	}
	return http.DetectContentType(data), nil
}

func getFileExt(name string) string {
	results := strings.Split(name, ".")
	if len(results) == 1 {
		return "undefined"
	}
	return results[len(results)-1]
}

const (
	pic   = "pic"
	video = "video"
	doc   = "doc"
)

// 微信匹配文件类型策略
func getMessageType(filename string) string {
	ext := getFileExt(filename)
	if imageType[ext] {
		return pic
	} else if ext == videoType {
		return video
	}
	return doc
}

func scanXml(resp *http.Response, v interface{}) error {
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(resp.Body); err != nil {
		return err
	}
	return xml.Unmarshal(buffer.Bytes(), v)
}

func scanJson(resp *http.Response, v interface{}) error {
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(resp.Body); err != nil {
		return err
	}
	return json.Unmarshal(buffer.Bytes(), v)
}

func stringToByte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&*(*reflect.StringHeader)(unsafe.Pointer(&s))))
}
