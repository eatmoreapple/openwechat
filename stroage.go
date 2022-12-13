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
	UUID         string
}

// HotReloadStorage 热登陆存储接口
type HotReloadStorage io.ReadWriter

// jsonFileHotReloadStorage 实现HotReloadStorage接口
// 默认以json文件的形式存储
type jsonFileHotReloadStorage struct {
	filename string
	file     *os.File
}

func (j *jsonFileHotReloadStorage) Read(p []byte) (n int, err error) {
	if j.file == nil {
		j.file, err = os.OpenFile(j.filename, os.O_RDWR, 0600)
		if os.IsNotExist(err) {
			return 0, ErrInvalidStorage
		}
		if err != nil {
			return 0, err
		}
	}
	return j.file.Read(p)
}

func (j *jsonFileHotReloadStorage) Write(p []byte) (n int, err error) {
	if j.file == nil {
		j.file, err = os.Create(j.filename)
		if err != nil {
			return 0, err
		}
	}
	j.file.Seek(0, io.SeekStart)
	if err = j.file.Truncate(0); err != nil {
		return
	}
	return j.file.Write(p)
}

func (j *jsonFileHotReloadStorage) Close() error {
	if j.file == nil {
		return nil
	}
	return j.file.Close()
}

// NewJsonFileHotReloadStorage 创建JsonFileHotReloadStorage
func NewJsonFileHotReloadStorage(filename string) io.ReadWriteCloser {
	return &jsonFileHotReloadStorage{filename: filename}
}

var _ HotReloadStorage = (*jsonFileHotReloadStorage)(nil)
