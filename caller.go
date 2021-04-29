package openwechat

import (
    "errors"
    "fmt"
    "os"
)

// 调用请求和解析请求
// 上层模块可以直接获取封装后的请求结果
type Caller struct {
    Client *Client
}

// Constructor for Caller
func NewCaller(client *Client) *Caller {
    return &Caller{Client: client}
}

// Default Constructor for Caller
func DefaultCaller(urlManager UrlManager) *Caller {
    return NewCaller(DefaultClient(urlManager))
}

// 获取登录的uuid
func (c *Caller) GetLoginUUID() (string, error) {
    resp := NewReturnResponse(c.Client.GetLoginUUID())
    if resp.Err() != nil {
        return "", resp.Err()
    }
    defer resp.Body.Close()
    data, err := resp.ReadAll()
    if err != nil {
        return "", err
    }
    // 正则匹配uuid字符串
    results := uuidRegexp.FindSubmatch(data)
    if len(results) != 2 {
        // 如果没有匹配到,可能微信的接口做了修改，或者当前机器的ip被加入了黑名单
        return "", errors.New("uuid does not match")
    }
    return string(results[1]), nil
}

// 检查是否登录成功
func (c *Caller) CheckLogin(uuid string) (*CheckLoginResponse, error) {
    resp := NewReturnResponse(c.Client.CheckLogin(uuid))
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    defer resp.Body.Close()
    data, err := resp.ReadAll()
    if err != nil {
        return nil, err
    }
    // 正则匹配检测的code
    // 具体code参考global.go
    results := statusCodeRegexp.FindSubmatch(data)
    if len(results) != 2 {
        return nil, errors.New("error status code match")
    }
    code := string(results[1])
    return &CheckLoginResponse{Code: code, Raw: data}, nil
}

// 获取登录信息
func (c *Caller) GetLoginInfo(body []byte) (*LoginInfo, error) {
    // 从响应体里面获取需要跳转的url
    results := redirectUriRegexp.FindSubmatch(body)
    if len(results) != 2 {
        return nil, errors.New("redirect url does not match")
    }
    path := string(results[1])
    resp := NewReturnResponse(c.Client.GetLoginInfo(path))
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    defer resp.Body.Close()
    var loginInfo LoginInfo
    // xml结构体序列化储存
    if err := resp.ScanXML(&loginInfo); err != nil {
        return nil, err
    }
    if !loginInfo.Ok() {
        return nil, loginInfo
    }
    return &loginInfo, nil
}

// 获取初始化信息
func (c *Caller) WebInit(request *BaseRequest) (*WebInitResponse, error) {
    resp := NewReturnResponse(c.Client.WebInit(request))
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    var webInitResponse WebInitResponse
    defer resp.Body.Close()
    if err := resp.ScanJSON(&webInitResponse); err != nil {
        return nil, err
    }
    return &webInitResponse, nil
}

// 通知手机已登录
func (c *Caller) WebWxStatusNotify(request *BaseRequest, response *WebInitResponse, info *LoginInfo) error {
    resp := NewReturnResponse(c.Client.WebWxStatusNotify(request, response, info))
    if resp.Err() != nil {
        return resp.Err()
    }
    var item struct{ BaseResponse BaseResponse }
    defer resp.Body.Close()
    if err := resp.ScanJSON(&item); err != nil {
        return err
    }
    if !item.BaseResponse.Ok() {
        return item.BaseResponse
    }
    return nil
}

// 异步获取是否有新的消息
func (c *Caller) SyncCheck(info *LoginInfo, response *WebInitResponse) (*SyncCheckResponse, error) {
    resp := NewReturnResponse(c.Client.SyncCheck(info, response))
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    defer resp.Body.Close()
    data, err := resp.ReadAll()
    if err != nil {
        return nil, err
    }
    results := syncCheckRegexp.FindSubmatch(data)
    if len(results) != 3 {
        return nil, errors.New("parse sync key failed")
    }
    retCode, selector := string(results[1]), string(results[2])
    syncCheckResponse := &SyncCheckResponse{RetCode: retCode, Selector: selector}
    return syncCheckResponse, nil
}

// 获取所有的联系人
func (c *Caller) WebWxGetContact(info *LoginInfo) (Members, error) {
    resp := NewReturnResponse(c.Client.WebWxGetContact(info))
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    defer resp.Body.Close()
    var item WebWxContactResponse
    if err := resp.ScanJSON(&item); err != nil {
        return nil, err
    }
    if !item.BaseResponse.Ok() {
        return nil, item.BaseResponse
    }
    return item.MemberList, nil
}

// 获取联系人的详情
// 注: Members参数的长度不要大于50
func (c *Caller) WebWxBatchGetContact(members Members, request *BaseRequest) (Members, error) {
    resp := NewReturnResponse(c.Client.WebWxBatchGetContact(members, request))
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    defer resp.Body.Close()
    var item WebWxBatchContactResponse
    if err := resp.ScanJSON(&item); err != nil {
        return nil, err
    }
    if !item.BaseResponse.Ok() {
        return nil, item.BaseResponse
    }
    return item.ContactList, nil
}

// 获取新的消息接口
func (c *Caller) WebWxSync(request *BaseRequest, response *WebInitResponse, info *LoginInfo) (*WebWxSyncResponse, error) {
    resp := NewReturnResponse(c.Client.WebWxSync(request, response, info))
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    defer resp.Body.Close()
    var webWxSyncResponse WebWxSyncResponse
    if err := resp.ScanJSON(&webWxSyncResponse); err != nil {
        return nil, err
    }
    return &webWxSyncResponse, nil
}

// 发送消息接口
func (c *Caller) WebWxSendMsg(msg *SendMessage, info *LoginInfo, request *BaseRequest) (*SentMessage, error) {
    resp := NewReturnResponse(c.Client.WebWxSendMsg(msg, info, request))
    sendSuccessMsg := &SentMessage{SendMessage: msg}
    err := parseMessageResponseError(resp, sendSuccessMsg)
    return sendSuccessMsg, err
}

// 修改用户备注接口
func (c *Caller) WebWxOplog(request *BaseRequest, remarkName, toUserName string) error {
    resp := NewReturnResponse(c.Client.WebWxOplog(request, remarkName, toUserName))
    return parseBaseResponseError(resp)
}

// 发送图片消息接口
func (c *Caller) WebWxSendImageMsg(file *os.File, request *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*SentMessage, error) {
    // 首先尝试上传图片
    sate, err := file.Stat()
    if err != nil {
        return nil, err
    }

    var resp *ReturnResponse
    if sate.Size() <= chunkSize {
        resp = NewReturnResponse(c.Client.WebWxUploadMedia(file, request, info, fromUserName, toUserName, "pic"))
    } else {
        resp = NewReturnResponse(c.Client.WebWxUploadMediaByChunk(file, request, info, fromUserName, toUserName, "pic"))
    }
    // 无错误上传成功之后获取请求结果，判断结果是否正常
    if resp.Err() != nil {
        return nil, resp.Err()
    }
    defer resp.Body.Close()
    data, _ := resp.ReadAll()
    fmt.Println(string(data))
    var item struct {
        BaseResponse BaseResponse
        MediaId      string
    }

    if err = resp.ScanJSON(&item); err != nil {
        return nil, err
    }
    if !item.BaseResponse.Ok() {
        return nil, item.BaseResponse
    }
    if len(item.MediaId) == 0 {
        return nil, errors.New("upload failed")
    }
    // 构造新的图片类型的信息
    msg := NewMediaSendMessage(ImageMessage, fromUserName, toUserName, item.MediaId)
    // 发送图片信息
    resp = NewReturnResponse(c.Client.WebWxSendMsgImg(msg, request, info))

    sendSuccessMsg := &SentMessage{SendMessage: msg}
    err = parseMessageResponseError(resp, sendSuccessMsg)
    return sendSuccessMsg, err
}

// 用户退出
func (c *Caller) Logout(info *LoginInfo) error {
    resp := NewReturnResponse(c.Client.Logout(info))
    return parseBaseResponseError(resp)
}

// 拉好友入群
func (c *Caller) AddFriendIntoChatRoom(req *BaseRequest, info *LoginInfo, group *Group, friends ...*Friend) error {
    if len(friends) == 0 {
        return errors.New("no friends found")
    }
    resp := NewReturnResponse(c.Client.AddMemberIntoChatRoom(req, info, group, friends...))
    return parseBaseResponseError(resp)
}

// 从群聊中移除用户
func (c *Caller) RemoveFriendFromChatRoom(req *BaseRequest, info *LoginInfo, group *Group, users ...*User) error {
    if len(users) == 0 {
        return errors.New("no users found")
    }
    resp := NewReturnResponse(c.Client.RemoveMemberFromChatRoom(req, info, group, users...))
    return parseBaseResponseError(resp)
}

// 同意加好友请求
func (c *Caller) WebWxVerifyUser(storage *Storage, info RecommendInfo, verifyContent string) error {
    resp := NewReturnResponse(c.Client.WebWxVerifyUser(storage, info, verifyContent))
    return parseBaseResponseError(resp)
}

// 撤回消息操作
func (c *Caller) WebWxRevokeMsg(msg *SentMessage, request *BaseRequest) error {
    resp := NewReturnResponse(c.Client.WebWxRevokeMsg(msg, request))
    return parseBaseResponseError(resp)
}

// 处理响应返回的结果是否正常
func parseBaseResponseError(resp *ReturnResponse) error {
    if resp.Err() != nil {
        return resp.Err()
    }
    defer resp.Body.Close()
    var item struct{ BaseResponse BaseResponse }
    if err := resp.ScanJSON(&item); err != nil {
        return err
    }
    if !item.BaseResponse.Ok() {
        return item.BaseResponse
    }
    return nil
}

func parseMessageResponseError(resp *ReturnResponse, msg *SentMessage) error {
    if resp.Err() != nil {
        return resp.Err()
    }

    defer resp.Body.Close()

    var messageResp MessageResponse

    if err := resp.ScanJSON(&messageResp); err != nil {
        return err
    }

    if !messageResp.BaseResponse.Ok() {
        return messageResp.BaseResponse
    }
    //// 发送成功之后将msgId赋值给SendMessage
    msg.MsgId = messageResp.MsgID
    return nil
}
