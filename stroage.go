package openwechat

import (
	"io"
	"net/http"
	"os"
)

// Storage 身份信息, 维持整个登陆的Session会话
type Storage struct {
	LoginInfo *LoginInfo
	Request   *BaseRequest
	Response  *WebInitResponse
}

type HotReloadStorageItem struct {
	Cookies      map[string][]*http.Cookie
	BaseRequest  *BaseRequest
	LoginInfo    *LoginInfo
	WechatDomain WechatDomain
}

// HotReloadStorage 热登陆存储接口
type HotReloadStorage io.ReadWriter

// JsonFileHotReloadStorage 实现HotReloadStorage接口
// 默认以json文件的形式存储
type JsonFileHotReloadStorage struct {
	FileName string
	file     *os.File
}

func (j *JsonFileHotReloadStorage) Read(p []byte) (n int, err error) {
	if j.file == nil {
		j.file, err = os.Open(j.FileName)
		if err != nil {
			return 0, err
		}
	}
	n, err = j.file.Read(p)
	if err == io.EOF {
		j.file.Close()
	}
	return n, err
}

func (j *JsonFileHotReloadStorage) Write(p []byte) (n int, err error) {
	file, err := os.Create(j.FileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(p)
}

func NewJsonFileHotReloadStorage(filename string) *JsonFileHotReloadStorage {
	return &JsonFileHotReloadStorage{FileName: filename}
}
