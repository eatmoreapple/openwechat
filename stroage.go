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

// 热登陆存储接口
type HotReloadStorage interface {
    GetCookie() map[string][]*http.Cookie                                            // 获取client.cookie
    GetBaseRequest() *BaseRequest                                                    // 获取BaseRequest
    GetLoginInfo() *LoginInfo                                                        // 获取LoginInfo
    Dump(cookies map[string][]*http.Cookie, req *BaseRequest, info *LoginInfo) error // 实现该方法, 将必要信息进行序列化
    Load() error                                                                     // 实现该方法, 将存储媒介的内容反序列化
}

// 实现HotReloadStorage接口
// 默认以json文件的形式存储
type JsonFileHotReloadStorage struct {
    Cookie   map[string][]*http.Cookie
    Req      *BaseRequest
    Info     *LoginInfo
    filename string
}

// 将信息写入json文件
func (f *JsonFileHotReloadStorage) Dump(cookies map[string][]*http.Cookie, req *BaseRequest, info *LoginInfo) error {
    var (
        file *os.File
        err  error
    )
    _, err = os.Stat(f.filename)
    if err != nil {
        if os.IsNotExist(err) {
            file, err = os.Create(f.filename)
            if err != nil {
                return err
            }
        }
    }

    if file == nil {
        file, err = os.Open(f.filename)
    }

    if err != nil {
        return err
    }
    defer file.Close()

    f.Cookie = cookies
    f.Req = req
    f.Info = info

    data, err := json.Marshal(f)
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
    return json.Unmarshal(buffer.Bytes(), f)
}

func (f *JsonFileHotReloadStorage) GetCookie() map[string][]*http.Cookie {
    return f.Cookie
}

func (f *JsonFileHotReloadStorage) GetBaseRequest() *BaseRequest {
    return f.Req
}

func (f *JsonFileHotReloadStorage) GetLoginInfo() *LoginInfo {
    return f.Info
}

func NewJsonFileHotReloadStorage(filename string) *JsonFileHotReloadStorage {
    return &JsonFileHotReloadStorage{filename: filename}
}
