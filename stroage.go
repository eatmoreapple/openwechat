package openwechat

import (
    "bytes"
    "encoding/json"
    "net/http"
    "os"
)

type Storage struct {
    LoginInfo *LoginInfo
    Request   *BaseRequest
    Response  *WebInitResponse
}

type HotReloadStorage interface {
    GetCookie() []*http.Cookie
    GetBaseRequest() *BaseRequest
    GetLoginInfo() *LoginInfo
    Dump(cookies []*http.Cookie, req *BaseRequest, info *LoginInfo) error
    Load() error
}

type FileHotReloadStorage struct {
    Cookie   []*http.Cookie
    Req      *BaseRequest
    Info     *LoginInfo
    filename string
}

func (f *FileHotReloadStorage) Dump(cookies []*http.Cookie, req *BaseRequest, info *LoginInfo) error {
    f.Cookie = cookies
    f.Req = req
    f.Info = info
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

    data, err := json.Marshal(f)
    if err != nil {
        return err
    }
    _, err = file.Write(data)
    return err
}

func (f *FileHotReloadStorage) Load() error {
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

func (f FileHotReloadStorage) GetCookie() []*http.Cookie {
    return f.Cookie
}

func (f FileHotReloadStorage) GetBaseRequest() *BaseRequest {
    return f.Req
}

func (f FileHotReloadStorage) GetLoginInfo() *LoginInfo {
    return f.Info
}

func NewFileHotReloadStorage(filename string) *FileHotReloadStorage {
    return &FileHotReloadStorage{filename: filename}
}
