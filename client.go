package openwechat

import (
    "bufio"
    "bytes"
    "crypto/md5"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

// HttpHook 请求上下文钩子
type HttpHook interface {
    BeforeRequest(req *http.Request)
    AfterRequest(response *http.Response, err error)
}

type HttpHooks []HttpHook

type UserAgentHook struct{}

func (u UserAgentHook) BeforeRequest(req *http.Request) {
    req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")
}

func (u UserAgentHook) AfterRequest(response *http.Response, err error) {}

// Client http请求客户端
// 客户端需要维持Session会话
// 并且客户端不允许跳转
type Client struct {
    HttpHooks HttpHooks
    *http.Client
    Domain  WechatDomain
    mode    mode
    mu      sync.Mutex
    cookies map[string][]*http.Cookie
}

func NewClient(client *http.Client) *Client {
    return &Client{Client: client}
}

// DefaultClient 自动存储cookie
// 设置客户端不自动跳转
func DefaultClient() *Client {
    jar, _ := cookiejar.New(nil)
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
        Jar: jar,
    }
    c := NewClient(client)
    c.AddHttpHook(UserAgentHook{})
    return c
}

func (c *Client) AddHttpHook(hooks ...HttpHook) {
    c.HttpHooks = append(c.HttpHooks, hooks...)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
    for _, hook := range c.HttpHooks {
        hook.BeforeRequest(req)
    }
    resp, err := c.Client.Do(req)
    for _, hook := range c.HttpHooks {
        hook.AfterRequest(resp, err)
    }
    return resp, err
}

func (c *Client) setCookie(resp *http.Response) {
    c.mu.Lock()
    defer c.mu.Unlock()
    cookies := resp.Cookies()
    if c.cookies == nil {
        c.cookies = make(map[string][]*http.Cookie)
    }
    path := fmt.Sprintf("%s://%s%s", resp.Request.URL.Scheme, resp.Request.URL.Host, resp.Request.URL.Path)
    c.cookies[path] = cookies
}

// Do 抽象Do方法,将所有的有效的cookie存入Client.cookies
// 方便热登陆时获取
func (c *Client) Do(req *http.Request) (*http.Response, error) {
    resp, err := c.do(req)
    if err == nil {
        c.setCookie(resp)
    }
    return resp, err
}

// GetCookieMap 获取当前client的所有的有效的client
func (c *Client) GetCookieMap() map[string][]*http.Cookie {
    return c.cookies
}

// GetLoginUUID 获取登录的uuid
func (c *Client) GetLoginUUID() (*http.Response, error) {
    path, _ := url.Parse(jslogin)
    params := url.Values{}
    redirectUrl, _ := url.Parse(webwxnewloginpage)
    if c.mode == Desktop {
        p := url.Values{"mod": {"desktop"}}
        redirectUrl.RawQuery = p.Encode()
    }
    params.Add("redirect_uri", redirectUrl.String())
    params.Add("appid", appId)
    params.Add("fun", "new")
    params.Add("lang", "zh_CN")
    params.Add("_", strconv.FormatInt(time.Now().Unix(), 10))

    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    return c.Do(req)
}

// GetLoginQrcode 获取登录的二维吗
func (c *Client) GetLoginQrcode(uuid string) (*http.Response, error) {
    path := qrcode + uuid
    return c.Get(path)
}

// CheckLogin 检查是否登录
func (c *Client) CheckLogin(uuid string) (*http.Response, error) {
    path, _ := url.Parse(login)
    now := time.Now().Unix()
    params := url.Values{}
    params.Add("r", strconv.FormatInt(now/1579, 10))
    params.Add("_", strconv.FormatInt(now, 10))
    params.Add("loginicon", "true")
    params.Add("uuid", uuid)
    params.Add("tip", "0")
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    return c.Do(req)
}

// GetLoginInfo 请求获取LoginInfo
func (c *Client) GetLoginInfo(path string) (*http.Response, error) {
    req, _ := http.NewRequest(http.MethodGet, path, nil)
    if c.mode == Desktop {
        req.Header.Add("client-version", uosPatchClientVersion)
        req.Header.Add("extspam", uosPatchExtspam)
    }
    return c.Do(req)
}

// WebInit 请求获取初始化信息
func (c *Client) WebInit(request *BaseRequest) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxinit)
    params := url.Values{}
    params.Add("_", fmt.Sprintf("%d", time.Now().Unix()))
    path.RawQuery = params.Encode()
    content := struct{ BaseRequest *BaseRequest }{BaseRequest: request}
    body, err := ToBuffer(content)
    if err != nil {
        return nil, err
    }
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// WebWxStatusNotify 通知手机已登录
func (c *Client) WebWxStatusNotify(request *BaseRequest, response *WebInitResponse, info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxstatusnotify)
    params := url.Values{}
    params.Add("lang", "zh_CN")
    params.Add("pass_ticket", info.PassTicket)
    username := response.User.UserName
    content := map[string]interface{}{
        "BaseRequest":  request,
        "ClientMsgId":  time.Now().Unix(),
        "Code":         3,
        "FromUserName": username,
        "ToUserName":   username,
    }
    path.RawQuery = params.Encode()
    buffer, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, path.String(), buffer)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// SyncCheck 异步检查是否有新的消息返回
func (c *Client) SyncCheck(info *LoginInfo, response *WebInitResponse) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.SyncHost() + synccheck)
    params := url.Values{}
    params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
    params.Add("skey", info.SKey)
    params.Add("sid", info.WxSid)
    params.Add("uin", strconv.Itoa(info.WxUin))
    params.Add("deviceid", GetRandomDeviceId())
    params.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
    var syncKeyStringSlice = make([]string, response.SyncKey.Count)
    // 将SyncKey里面的元素按照特定的格式拼接起来
    for index, item := range response.SyncKey.List {
        i := fmt.Sprintf("%d_%d", item.Key, item.Val)
        syncKeyStringSlice[index] = i
    }
    syncKey := strings.Join(syncKeyStringSlice, "|")
    params.Add("synckey", syncKey)
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    return c.Do(req)
}

// WebWxGetContact 获取联系人信息
func (c *Client) WebWxGetContact(info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxgetcontact)
    params := url.Values{}
    params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
    params.Add("skey", info.SKey)
    params.Add("req", "0")
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    return c.Do(req)
}

// WebWxBatchGetContact 获取联系人详情
func (c *Client) WebWxBatchGetContact(members Members, request *BaseRequest) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxbatchgetcontact)
    params := url.Values{}
    params.Add("type", "ex")
    params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
    path.RawQuery = params.Encode()
    list := NewUserDetailItemList(members)
    content := map[string]interface{}{
        "BaseRequest": request,
        "Count":       members.Count(),
        "List":        list,
    }
    body, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// WebWxSync 获取消息接口
func (c *Client) WebWxSync(request *BaseRequest, response *WebInitResponse, info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxsync)
    params := url.Values{}
    params.Add("sid", info.WxSid)
    params.Add("skey", info.SKey)
    params.Add("pass_ticket", info.PassTicket)
    path.RawQuery = params.Encode()
    content := map[string]interface{}{
        "BaseRequest": request,
        "SyncKey":     response.SyncKey,
        "rr":          strconv.FormatInt(time.Now().Unix(), 10),
    }
    data, _ := json.Marshal(content)
    body := bytes.NewBuffer(data)
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// 发送消息
func (c *Client) sendMessage(request *BaseRequest, url string, msg *SendMessage) (*http.Response, error) {
    content := map[string]interface{}{
        "BaseRequest": request,
        "Msg":         msg,
        "Scene":       0,
    }
    body, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, url, body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// WebWxSendMsg 发送文本消息
func (c *Client) WebWxSendMsg(msg *SendMessage, info *LoginInfo, request *BaseRequest) (*http.Response, error) {
    msg.Type = MsgTypeText
    path, _ := url.Parse(c.Domain.BaseHost() + webwxsendmsg)
    params := url.Values{}
    params.Add("lang", "zh_CN")
    params.Add("pass_ticket", info.PassTicket)
    path.RawQuery = params.Encode()
    return c.sendMessage(request, path.String(), msg)
}

// WebWxGetHeadImg 获取用户的头像
func (c *Client) WebWxGetHeadImg(headImageUrl string) (*http.Response, error) {
    path := c.Domain.BaseHost() + headImageUrl
    req, _ := http.NewRequest(http.MethodGet, path, nil)
    return c.Do(req)
}

func (c *Client) WebWxUploadMediaByChunk(file *os.File, request *BaseRequest, info *LoginInfo, forUserName, toUserName string) (*http.Response, error) {
    // 获取文件上传的类型
    contentType, err := GetFileContentType(file)
    if err != nil {
        return nil, err
    }

    // 将文件的游标复原到原点
    // 上面获取文件的类型的时候已经读取了512个字节
    if _, err = file.Seek(0, 0); err != nil {
        return nil, err
    }

    reader := bufio.NewReader(file)

    h := md5.New()
    if _, err = io.Copy(h, reader); err != nil {
        return nil, err
    }

    fileMd5 := fmt.Sprintf("%x", h.Sum(nil))

    sate, err := file.Stat()
    if err != nil {
        return nil, err
    }

    // 获取文件的类型
    mediaType := getMessageType(sate.Name())

    path, _ := url.Parse(c.Domain.FileHost() + webwxuploadmedia)
    params := url.Values{}
    params.Add("f", "json")

    path.RawQuery = params.Encode()

    cookies := c.Jar.Cookies(path)
    webWxDataTicket := getWebWxDataTicket(cookies)

    uploadMediaRequest := map[string]interface{}{
        "UploadType":    2,
        "BaseRequest":   request,
        "ClientMediaId": time.Now().Unix() * 1e4,
        "TotalLen":      sate.Size(),
        "StartPos":      0,
        "DataLen":       sate.Size(),
        "MediaType":     4,
        "FromUserName":  forUserName,
        "ToUserName":    toUserName,
        "FileMd5":       fileMd5,
    }

    uploadMediaRequestByte, err := json.Marshal(uploadMediaRequest)
    if err != nil {
        return nil, err
    }

    var chunks int64

    if sate.Size() > chunkSize {
        chunks = sate.Size() / chunkSize
        if chunks*chunkSize < sate.Size() {
            chunks++
        }
    } else {
        chunks = 1
    }

    var resp *http.Response

    content := map[string]string{
        "id":                "WU_FILE_0",
        "name":              file.Name(),
        "type":              contentType,
        "lastModifiedDate":  sate.ModTime().Format(TimeFormat),
        "size":              strconv.FormatInt(sate.Size(), 10),
        "mediatype":         mediaType,
        "webwx_data_ticket": webWxDataTicket,
        "pass_ticket":       info.PassTicket,
    }

    if chunks > 1 {
        content["chunks"] = strconv.FormatInt(chunks, 10)
    }

    if _, err = file.Seek(0, 0); err != nil {
        return nil, err
    }

    // 分块上传
    for chunk := 0; int64(chunk) < chunks; chunk++ {

        var isLastTime bool
        if int64(chunk)+1 == chunks {
            isLastTime = true
        }

        if chunks > 1 {
            content["chunk"] = strconv.Itoa(chunk)
        }

        var formBuffer bytes.Buffer

        writer := multipart.NewWriter(&formBuffer)
        if err = writer.WriteField("uploadmediarequest", string(uploadMediaRequestByte)); err != nil {
            return nil, err
        }

        for k, v := range content {
            if err := writer.WriteField(k, v); err != nil {
                return nil, err
            }
        }

        if w, err := writer.CreateFormFile("filename", file.Name()); err != nil {
            return nil, err
        } else {
            chunkData := make([]byte, chunkSize)
            if _, err := file.Read(chunkData); err != nil && err != io.EOF {
                return nil, err
            }
            if _, err = w.Write(chunkData); err != nil {
                return nil, err
            }
        }
        ct := writer.FormDataContentType()
        if err = writer.Close(); err != nil {
            return nil, err
        }
        req, _ := http.NewRequest(http.MethodPost, path.String(), &formBuffer)
        req.Header.Set("Content-Type", ct)
        // 发送数据
        resp, err = c.Do(req)
        if err != nil {
            return nil, err
        }
        // 如果不是最后一次, 解析有没有错误
        if !isLastTime {
            if err := parseBaseResponseError(resp); err != nil {
                return nil, err
            }
        }
    }
    // 将最后一次携带文件信息的response返回
    return resp, err
}

// WebWxSendMsgImg 发送图片
// 这个接口依赖上传文件的接口
// 发送的图片必须是已经成功上传的图片
func (c *Client) WebWxSendMsgImg(msg *SendMessage, request *BaseRequest, info *LoginInfo) (*http.Response, error) {
    msg.Type = MsgTypeImage
    path, _ := url.Parse(c.Domain.BaseHost() + webwxsendmsgimg)
    params := url.Values{}
    params.Add("fun", "async")
    params.Add("f", "json")
    params.Add("lang", "zh_CN")
    params.Add("pass_ticket", info.PassTicket)
    path.RawQuery = params.Encode()
    return c.sendMessage(request, path.String(), msg)
}

// WebWxSendAppMsg 发送文件信息
func (c *Client) WebWxSendAppMsg(msg *SendMessage, request *BaseRequest) (*http.Response, error) {
    msg.Type = AppMessage
    path, _ := url.Parse(c.Domain.BaseHost() + webwxsendappmsg)
    params := url.Values{}
    params.Add("fun", "async")
    params.Add("f", "json")
    path.RawQuery = params.Encode()
    return c.sendMessage(request, path.String(), msg)
}

// WebWxOplog 用户重命名接口
func (c *Client) WebWxOplog(request *BaseRequest, remarkName, userName string) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxoplog)
    params := url.Values{}
    params.Add("lang", "zh_CN")
    path.RawQuery = params.Encode()
    content := map[string]interface{}{
        "BaseRequest": request,
        "CmdId":       2,
        "RemarkName":  remarkName,
        "UserName":    userName,
    }
    body, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// WebWxVerifyUser 添加用户为好友接口
func (c *Client) WebWxVerifyUser(storage *Storage, info RecommendInfo, verifyContent string) (*http.Response, error) {
    loginInfo := storage.LoginInfo
    path, _ := url.Parse(c.Domain.BaseHost() + webwxverifyuser)
    params := url.Values{}
    params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
    params.Add("lang", "zh_CN")
    params.Add("pass_ticket", loginInfo.PassTicket)
    path.RawQuery = params.Encode()
    content := map[string]interface{}{
        "BaseRequest":    storage.Request,
        "Opcode":         3,
        "SceneList":      [1]int{33},
        "SceneListCount": 1,
        "VerifyContent":  verifyContent,
        "VerifyUserList": []interface{}{map[string]string{
            "Value":            info.UserName,
            "VerifyUserTicket": info.Ticket,
        }},
        "VerifyUserListSize": 1,
        "skey":               loginInfo.SKey,
    }
    body, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// WebWxGetMsgImg 获取图片消息的图片响应
func (c *Client) WebWxGetMsgImg(msg *Message, info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxgetmsgimg)
    params := url.Values{}
    params.Add("MsgID", msg.MsgId)
    params.Add("skey", info.SKey)
    // params.Add("type", "slave")
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    return c.Do(req)
}

// WebWxGetVoice 获取语音消息的语音响应
func (c *Client) WebWxGetVoice(msg *Message, info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxgetvoice)
    params := url.Values{}
    params.Add("msgid", msg.MsgId)
    params.Add("skey", info.SKey)
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    req.Header.Add("Referer", path.String())
    req.Header.Add("Range", "bytes=0-")
    return c.Do(req)
}

// WebWxGetVideo 获取视频消息的视频响应
func (c *Client) WebWxGetVideo(msg *Message, info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxgetvideo)
    params := url.Values{}
    params.Add("msgid", msg.MsgId)
    params.Add("skey", info.SKey)
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    req.Header.Add("Referer", path.String())
    req.Header.Add("Range", "bytes=0-")
    return c.Do(req)
}

// WebWxGetMedia 获取文件消息的文件响应
func (c *Client) WebWxGetMedia(msg *Message, info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.FileHost() + webwxgetmedia)
    params := url.Values{}
    params.Add("sender", msg.FromUserName)
    params.Add("mediaid", msg.MediaId)
    params.Add("encryfilename", msg.EncryFileName)
    params.Add("fromuser", fmt.Sprintf("%d", info.WxUin))
    params.Add("pass_ticket", info.PassTicket)
    params.Add("webwx_data_ticket", getWebWxDataTicket(c.Jar.Cookies(path)))
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    req.Header.Add("Referer", path.String())
    req.Header.Add("Range", "bytes=0-")
    return c.Do(req)
}

// Logout 用户退出
func (c *Client) Logout(info *LoginInfo) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxlogout)
    params := url.Values{}
    params.Add("redirect", "1")
    params.Add("type", "1")
    params.Add("skey", info.SKey)
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    return c.Do(req)
}

// AddMemberIntoChatRoom 添加用户进群聊
func (c *Client) AddMemberIntoChatRoom(req *BaseRequest, info *LoginInfo, group *Group, friends ...*Friend) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxupdatechatroom)
    params := url.Values{}
    params.Add("fun", "addmember")
    params.Add("pass_ticket", info.PassTicket)
    params.Add("lang", "zh_CN")
    path.RawQuery = params.Encode()
    addMemberList := make([]string, len(friends))
    for index, friend := range friends {
        addMemberList[index] = friend.UserName
    }
    content := map[string]interface{}{
        "ChatRoomName":  group.UserName,
        "BaseRequest":   req,
        "AddMemberList": strings.Join(addMemberList, ","),
    }
    buffer, _ := ToBuffer(content)
    requ, _ := http.NewRequest(http.MethodPost, path.String(), buffer)
    requ.Header.Set("Content-Type", jsonContentType)
    return c.Do(requ)
}

// RemoveMemberFromChatRoom 从群聊中移除用户
func (c *Client) RemoveMemberFromChatRoom(req *BaseRequest, info *LoginInfo, group *Group, friends ...*User) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxupdatechatroom)
    params := url.Values{}
    params.Add("fun", "delmember")
    params.Add("lang", "zh_CN")
    params.Add("pass_ticket", info.PassTicket)
    delMemberList := make([]string, len(friends))
    for index, friend := range friends {
        delMemberList[index] = friend.UserName
    }
    content := map[string]interface{}{
        "ChatRoomName":  group.UserName,
        "BaseRequest":   req,
        "DelMemberList": strings.Join(delMemberList, ","),
    }
    buffer, _ := ToBuffer(content)
    requ, _ := http.NewRequest(http.MethodPost, path.String(), buffer)
    requ.Header.Set("Content-Type", jsonContentType)
    return c.Do(requ)
}

// WebWxRevokeMsg 撤回消息
func (c *Client) WebWxRevokeMsg(msg *SentMessage, request *BaseRequest) (*http.Response, error) {
    content := map[string]interface{}{
        "BaseRequest": request,
        "ClientMsgId": msg.ClientMsgId,
        "SvrMsgId":    msg.MsgId,
        "ToUserName":  msg.ToUserName,
    }
    buffer, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, c.Domain.BaseHost()+webwxrevokemsg, buffer)
    req.Header.Set("Content-Type", jsonContentType)
    return c.Do(req)
}

// 校验上传文件
func (c *Client) webWxCheckUpload(stat os.FileInfo, request *BaseRequest, fileMd5, fromUserName, toUserName string) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxcheckupload)
    content := map[string]interface{}{
        "BaseRequest":  request,
        "FileMd5":      fileMd5,
        "FileName":     stat.Name(),
        "FileSize":     stat.Size(),
        "FileType":     7,
        "FromUserName": fromUserName,
        "ToUserName":   toUserName,
    }
    body, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

func (c *Client) WebWxStatusAsRead(request *BaseRequest, info *LoginInfo, msg *Message) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxstatusnotify)
    content := map[string]interface{}{
        "BaseRequest":  request,
        "DeviceID":     request.DeviceID,
        "Sid":          request.Sid,
        "Skey":         request.Skey,
        "Uin":          info.WxUin,
        "ClientMsgId":  time.Now().Unix(),
        "Code":         1,
        "FromUserName": msg.ToUserName,
        "ToUserName":   msg.FromUserName,
    }
    body, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// WebWxRelationPin 联系人置顶接口
func (c *Client) WebWxRelationPin(request *BaseRequest, op uint8, user *User) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxoplog)
    content := map[string]interface{}{
        "BaseRequest": request,
        "CmdId":       3,
        "OP":          op,
        "RemarkName":  user.RemarkName,
        "UserName":    user.UserName,
    }
    body, _ := ToBuffer(content)
    req, _ := http.NewRequest(http.MethodPost, path.String(), body)
    req.Header.Add("Content-Type", jsonContentType)
    return c.Do(req)
}

// WebWxPushLogin 免扫码登陆接口
func (c *Client) WebWxPushLogin(uin int) (*http.Response, error) {
    path, _ := url.Parse(c.Domain.BaseHost() + webwxpushloginurl)
    params := url.Values{"uin": {strconv.Itoa(uin)}}
    path.RawQuery = params.Encode()
    req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
    return c.Do(req)
}
