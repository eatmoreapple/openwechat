package openwechat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

// 身份信息, 维持整个登陆的Session会话
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

// 热登陆存储接口
type HotReloadStorage interface {
	GetHotReloadStorageItem() HotReloadStorageItem // 获取HotReloadStorageItem
	Dump(item HotReloadStorageItem) error          // 实现该方法, 将必要信息进行序列化
	Load() error                                   // 实现该方法, 将存储媒介的内容反序列化
}

// 实现HotReloadStorage接口
// 默认以json文件的形式存储
type JsonFileHotReloadStorage struct {
	item     HotReloadStorageItem
	filename string
}

// 将信息写入json文件
func (f *JsonFileHotReloadStorage) Dump(item HotReloadStorageItem) error {

	file, err := os.OpenFile(f.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}

	defer file.Close()

	f.item = item

	data, err := json.Marshal(f.item)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}

// 从文件中读取信息
func (f *JsonFileHotReloadStorage) Load() error {
	file, err := os.Open(f.filename)

	if err != nil {
		return err
	}
	defer file.Close()
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(file); err != nil {
		return err
	}
	err = json.Unmarshal(buffer.Bytes(), &f.item)
	return err
}

func (f *JsonFileHotReloadStorage) GetHotReloadStorageItem() HotReloadStorageItem {
	return f.item
}

func NewJsonFileHotReloadStorage(filename string) *JsonFileHotReloadStorage {
	return &JsonFileHotReloadStorage{filename: filename}
}
