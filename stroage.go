package openwechat

import (
	"encoding/json"
	"errors"
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
// Deprecated
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
	return j.file.Read(p)
}

func (j *JsonFileHotReloadStorage) Write(p []byte) (n int, err error) {
	j.file, err = os.Create(j.FileName)
	if err != nil {
		return 0, err
	}
	return j.file.Write(p)
}

// NewJsonFileHotReloadStorage 创建JsonFileHotReloadStorage
// Deprecated
// use os.File instead
func NewJsonFileHotReloadStorage(filename string) HotReloadStorage {
	return &JsonFileHotReloadStorage{FileName: filename}
}

var _ HotReloadStorage = &JsonFileHotReloadStorage{}

func NewHotReloadStorageItem(storage HotReloadStorage) (*HotReloadStorageItem, error) {
	if storage == nil {
		return nil, errors.New("storage can't be nil")
	}
	var item HotReloadStorageItem

	if err := json.NewDecoder(storage).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}
