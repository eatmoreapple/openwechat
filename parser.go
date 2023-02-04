package openwechat

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"
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

// GetFileContentType 获取文件上传的类型
func GetFileContentType(file multipart.File) (string, error) {
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

func scanXml(reader io.Reader, v interface{}) error {
	return xml.NewDecoder(reader).Decode(v)
}

func scanJson(reader io.Reader, v interface{}) error {
	return json.NewDecoder(reader).Decode(v)
}

func stringToByte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&*(*reflect.StringHeader)(unsafe.Pointer(&s))))
}
