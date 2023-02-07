package openwechat

import (
	"io"
	"os"
	"sync"
	"time"
)

// Storage 身份信息, 维持整个登陆的Session会话
type Storage struct {
	LoginInfo *LoginInfo
	Request   *BaseRequest
	Response  *WebInitResponse
}

type HotReloadStorageItem struct {
	Jar          *Jar
	BaseRequest  *BaseRequest
	LoginInfo    *LoginInfo
	WechatDomain WechatDomain
	SyncKey      *SyncKey
	UUID         string
}

// HotReloadStorage 热登陆存储接口
type HotReloadStorage io.ReadWriter

// fileHotReloadStorage 实现HotReloadStorage接口
// 以文件的形式存储
type fileHotReloadStorage struct {
	filename string
	file     *os.File
	lock     sync.Mutex
}

func (j *fileHotReloadStorage) Read(p []byte) (n int, err error) {
	j.lock.Lock()
	defer j.lock.Unlock()
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

func (j *fileHotReloadStorage) Write(p []byte) (n int, err error) {
	j.lock.Lock()
	defer j.lock.Unlock()
	if j.file == nil {
		j.file, err = os.Create(j.filename)
		if err != nil {
			return 0, err
		}
	}
	// reset offset and truncate file
	if _, err = j.file.Seek(0, io.SeekStart); err != nil {
		return
	}
	if err = j.file.Truncate(0); err != nil {
		return
	}
	// json decode only write once
	return j.file.Write(p)
}

func (j *fileHotReloadStorage) Close() error {
	j.lock.Lock()
	defer j.lock.Unlock()
	if j.file == nil {
		return nil
	}
	return j.file.Close()
}

// Deprecated: use NewFileHotReloadStorage instead
// 不再单纯以json的格式存储，支持了用户自定义序列化方式
func NewJsonFileHotReloadStorage(filename string) io.ReadWriteCloser {
	return NewFileHotReloadStorage(filename)
}

// NewFileHotReloadStorage implements HotReloadStorage
func NewFileHotReloadStorage(filename string) io.ReadWriteCloser {
	return &fileHotReloadStorage{filename: filename}
}

var _ HotReloadStorage = (*fileHotReloadStorage)(nil)

type HotReloadStorageSyncer struct {
	duration time.Duration
	bot      *Bot
}

// Sync 定时同步数据到登陆存储中
func (h *HotReloadStorageSyncer) Sync() error {
	if h.duration <= 0 {
		return nil
	}
	// 定时器
	ticker := time.NewTicker(h.duration)
	for {
		select {
		case <-ticker.C:
			// 每隔一段时间, 将数据同步到storage中
			if err := h.bot.DumpHotReloadStorage(); err != nil {
				return err
			}
		case <-h.bot.Context().Done():
			// 当Bot关闭的时候, 退出循环
			return nil
		}
	}
}

func NewHotReloadStorageSyncer(bot *Bot, duration time.Duration) *HotReloadStorageSyncer {
	return &HotReloadStorageSyncer{duration: duration, bot: bot}
}
