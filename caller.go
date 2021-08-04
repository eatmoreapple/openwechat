package openwechat

import (
    "bytes"
    "encoding/json"
    "errors"
    "net/http"
    "net/url"
    "os"
)

// Caller 调用请求和解析请求
// 上层模块可以直接获取封装后的请求结果
type Caller struct {
    Client *Client
    path   *url.URL
}

// NewCaller Constructor for Caller
func NewCaller(client *Client) *Caller {
    return &Caller{Client: client}
}

// DefaultCaller Default Constructor for Caller
func DefaultCaller() *Caller {
    return NewCaller(DefaultClient())
}

// GetLoginUUID 获取登录的uuid
func (c *Caller) GetLoginUUID() (string, error) {
    resp, err := c.Client.GetLoginUUID()
    if err != nil {
        return "", err
    }

    defer resp.Body.Close()

    var buffer bytes.Buffer
    if _, err := buffer.ReadFrom(resp.Body); err != nil {
        return "", err
    }
    // 正则匹配uuid字符串
    results := uuidRegexp.FindSubmatch(buffer.Bytes())
    if len(results) != 2 {
        // 如果没有匹配到,可能微信的接口做了修改，或者当前机器的ip被加入了黑名单
        return "", errors.New("uuid does not match")
    }
    return string(results[1]), nil
}

// CheckLogin 检查是否登录成功
func (c *Caller) CheckLogin(uuid string) (*CheckLoginResponse, error) {
    resp, err := c.Client.CheckLogin(uuid)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var buffer bytes.Buffer
    if _, err := buffer.ReadFrom(resp.Body); err != nil {
        return nil, err
    }
    // 正则匹配检测的code
    // 具体code参考global.go
    results := statusCodeRegexp.FindSubmatch(buffer.Bytes())
    if len(results) != 2 {
        return nil, errors.New("error status code match")
    }
    code := string(results[1])
    return &CheckLoginResponse{Code: code, Raw: buffer.Bytes()}, nil
}

// GetLoginInfo 获取登录信息
func (c *Caller) GetLoginInfo(body []byte) (*LoginInfo, error) {
    // 从响应体里面获取需要跳转的url
    results := redirectUriRegexp.FindSubmatch(body)
    if len(results) != 2 {
        return nil, errors.New("redirect url does not match")
    }
    path, err := url.Parse(string(results[1]))
    if err != nil {
        return nil, err
    }
    c.Client.Domain = WechatDomain(path.Host)
    resp, err := c.Client.GetLoginInfo(path.String())
    if err != nil {
        uErr, ok := err.(*url.Error)
        if ok && (uErr.Err.Error() == ErrMissLocationHeader.Error()) {
            return nil, ErrLoginForbiddenError
        }
        return nil, err
    }
    defer resp.Body.Close()

    var loginInfo LoginInfo
    // xml结构体序列化储存
    if err := scanXml(resp, &loginInfo); err != nil {
        return nil, err
    }
    if !loginInfo.Ok() {
        return nil, loginInfo
    }
    return &loginInfo, nil
}

// WebInit 获取初始化信息
func (c *Caller) WebInit(request *BaseRequest) (*WebInitResponse, error) {
    resp, err := c.Client.WebInit(request)
    if err != nil {
        return nil, err
    }
    var webInitResponse WebInitResponse
    defer resp.Body.Close()
    if err := scanJson(resp, &webInitResponse); err != nil {
        return nil, err
    }
    return &webInitResponse, nil
}

// WebWxStatusNotify 通知手机已登录
func (c *Caller) WebWxStatusNotify(request *BaseRequest, response *WebInitResponse, info *LoginInfo) error {
    resp, err := c.Client.WebWxStatusNotify(request, response, info)
    if err != nil {
        return err
    }
    var item struct{ BaseResponse BaseResponse }
    defer resp.Body.Close()
    if err := scanJson(resp, &item); err != nil {
        return err
    }
    if !item.BaseResponse.Ok() {
        return item.BaseResponse
    }
    return nil
}

// SyncCheck 异步获取是否有新的消息
func (c *Caller) SyncCheck(info *LoginInfo, response *WebInitResponse) (*SyncCheckResponse, error) {
    resp, err := c.Client.SyncCheck(info, response)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    var buffer bytes.Buffer
    if _, err := buffer.ReadFrom(resp.Body); err != nil {
        return nil, err
    }
    results := syncCheckRegexp.FindSubmatch(buffer.Bytes())
    if len(results) != 3 {
        return nil, errors.New("parse sync key failed")
    }
    retCode, selector := string(results[1]), string(results[2])
    syncCheckResponse := &SyncCheckResponse{RetCode: retCode, Selector: selector}
    return syncCheckResponse, nil
}

// WebWxGetContact 获取所有的联系人
func (c *Caller) WebWxGetContact(info *LoginInfo) (Members, error) {
    resp, err := c.Client.WebWxGetContact(info)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    var item WebWxContactResponse
    if err := scanJson(resp, &item); err != nil {
        return nil, err
    }
    if !item.BaseResponse.Ok() {
        return nil, item.BaseResponse
    }
    return item.MemberList, nil
}

// WebWxBatchGetContact 获取联系人的详情
// 注: Members参数的长度不要大于50
func (c *Caller) WebWxBatchGetContact(members Members, request *BaseRequest) (Members, error) {
    resp, err := c.Client.WebWxBatchGetContact(members, request)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    var item WebWxBatchContactResponse
    if err := scanJson(resp, &item); err != nil {
        return nil, err
    }
    if !item.BaseResponse.Ok() {
        return nil, item.BaseResponse
    }
    return item.ContactList, nil
}

// WebWxSync 获取新的消息接口
func (c *Caller) WebWxSync(request *BaseRequest, response *WebInitResponse, info *LoginInfo) (*WebWxSyncResponse, error) {
    resp, err := c.Client.WebWxSync(request, response, info)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    var webWxSyncResponse WebWxSyncResponse
    if err := scanJson(resp, &webWxSyncResponse); err != nil {
        return nil, err
    }
    return &webWxSyncResponse, nil
}

// WebWxSendMsg 发送消息接口
func (c *Caller) WebWxSendMsg(msg *SendMessage, info *LoginInfo, request *BaseRequest) (*SentMessage, error) {
    resp, err := c.Client.WebWxSendMsg(msg, info, request)
    return getSuccessSentMessage(msg, resp, err)
}

// WebWxOplog 修改用户备注接口
func (c *Caller) WebWxOplog(request *BaseRequest, remarkName, toUserName string) error {
    resp, err := c.Client.WebWxOplog(request, remarkName, toUserName)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

func (c *Caller) UploadMedia(file *os.File, request *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*UploadResponse, error) {
    // 首先尝试上传图片
    resp, err := c.Client.WebWxUploadMediaByChunk(file, request, info, fromUserName, toUserName)
    // 无错误上传成功之后获取请求结果，判断结果是否正常
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var item UploadResponse

    if err := scanJson(resp, &item); err != nil {
        return &item, err
    }
    if !item.BaseResponse.Ok() {
        return &item, item.BaseResponse
    }
    if len(item.MediaId) == 0 {
        return &item, errors.New("upload failed")
    }
    return &item, nil
}

// WebWxSendImageMsg 发送图片消息接口
func (c *Caller) WebWxSendImageMsg(file *os.File, request *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*SentMessage, error) {
    // 首先尝试上传图片
    resp, err := c.UploadMedia(file, request, info, fromUserName, toUserName)
    if err != nil {
        return nil, err
    }
    // 构造新的图片类型的信息
    msg := NewMediaSendMessage(MsgTypeImage, fromUserName, toUserName, resp.MediaId)
    // 发送图片信息
    resp1, err := c.Client.WebWxSendMsgImg(msg, request, info)
    return getSuccessSentMessage(msg, resp1, err)
}

func (c *Caller) WebWxSendFile(file *os.File, req *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*SentMessage, error) {
    resp, err := c.UploadMedia(file, req, info, fromUserName, toUserName)
    if err != nil {
        return nil, err
    }
    // 构造新的文件类型的信息
    stat, _ := file.Stat()
    appMsg := NewFileAppMessage(stat, resp.MediaId)
    content, err := appMsg.XmlByte()
    if err != nil {
        return nil, err
    }
    msg := NewSendMessage(AppMessage, string(content), fromUserName, toUserName, "")
    return c.WebWxSendAppMsg(msg, req)
}

// WebWxSendAppMsg 发送媒体消息
func (c *Caller) WebWxSendAppMsg(msg *SendMessage, req *BaseRequest) (*SentMessage, error) {
    resp, err := c.Client.WebWxSendAppMsg(msg, req)
    return getSuccessSentMessage(msg, resp, err)
}

// Logout 用户退出
func (c *Caller) Logout(info *LoginInfo) error {
    resp, err := c.Client.Logout(info)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

// AddFriendIntoChatRoom 拉好友入群
func (c *Caller) AddFriendIntoChatRoom(req *BaseRequest, info *LoginInfo, group *Group, friends ...*Friend) error {
    if len(friends) == 0 {
        return errors.New("no friends found")
    }
    resp, err := c.Client.AddMemberIntoChatRoom(req, info, group, friends...)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

// RemoveFriendFromChatRoom 从群聊中移除用户
func (c *Caller) RemoveFriendFromChatRoom(req *BaseRequest, info *LoginInfo, group *Group, users ...*User) error {
    if len(users) == 0 {
        return errors.New("no users found")
    }
    resp, err := c.Client.RemoveMemberFromChatRoom(req, info, group, users...)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

// WebWxVerifyUser 同意加好友请求
func (c *Caller) WebWxVerifyUser(storage *Storage, info RecommendInfo, verifyContent string) error {
    resp, err := c.Client.WebWxVerifyUser(storage, info, verifyContent)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

// WebWxRevokeMsg 撤回消息操作
func (c *Caller) WebWxRevokeMsg(msg *SentMessage, request *BaseRequest) error {
    resp, err := c.Client.WebWxRevokeMsg(msg, request)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

// WebWxStatusAsRead 将消息设置为已读
func (c *Caller) WebWxStatusAsRead(request *BaseRequest, info *LoginInfo, msg *Message) error {
    resp, err := c.Client.WebWxStatusAsRead(request, info, msg)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

// WebWxRelationPin 将联系人是否置顶
func (c *Caller) WebWxRelationPin(request *BaseRequest, user *User, op uint8) error {
    resp, err := c.Client.WebWxRelationPin(request, op, user)
    if err != nil {
        return err
    }
    return parseBaseResponseError(resp)
}

// WebWxPushLogin 免扫码登陆接口
func (c *Caller) WebWxPushLogin(uin int) (*PushLoginResponse, error) {
    resp, err := c.Client.WebWxPushLogin(uin)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    var item PushLoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
        return nil, err
    }
    return &item, nil
}

// 处理响应返回的结果是否正常
func parseBaseResponseError(resp *http.Response) error {
    defer resp.Body.Close()
    var item struct{ BaseResponse BaseResponse }
    if err := scanJson(resp, &item); err != nil {
        return err
    }
    if !item.BaseResponse.Ok() {
        return item.BaseResponse
    }
    return nil
}

func parseMessageResponseError(resp *http.Response, msg *SentMessage) error {
    defer resp.Body.Close()

    var messageResp MessageResponse

    if err := scanJson(resp, &messageResp); err != nil {
        return err
    }

    if !messageResp.BaseResponse.Ok() {
        return messageResp.BaseResponse
    }
    // 发送成功之后将msgId赋值给SendMessage
    msg.MsgId = messageResp.MsgID
    return nil
}

func getSuccessSentMessage(msg *SendMessage, resp *http.Response, err error) (*SentMessage, error) {
    if err != nil {
        return nil, err
    }
    sendSuccessMsg := &SentMessage{SendMessage: msg}
    err = parseMessageResponseError(resp, sendSuccessMsg)
    return sendSuccessMsg, err
}
