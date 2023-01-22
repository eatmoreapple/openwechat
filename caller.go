package openwechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// Caller 调用请求和解析请求
// 上层模块可以直接获取封装后的请求结果
type Caller struct {
	Client *Client
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

	defer func() { _ = resp.Body.Close() }()

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
func (c *Caller) CheckLogin(uuid, tip string) (CheckLoginResponse, error) {
	resp, err := c.Client.CheckLogin(uuid, tip)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// GetLoginInfo 获取登录信息
func (c *Caller) GetLoginInfo(path *url.URL) (*LoginInfo, error) {
	// 从响应体里面获取需要跳转的url
	resp, err := c.Client.GetLoginInfo(path)
	if err != nil {
		return nil, err
	}
	// 判断是否重定向
	if resp.StatusCode != http.StatusMovedPermanently {
		return nil, fmt.Errorf("%w: try to login with Desktop Mode", ErrForbidden)
	}
	defer func() { _ = resp.Body.Close() }()

	var loginInfo LoginInfo
	// xml结构体序列化储存
	if err := scanXml(resp.Body, &loginInfo); err != nil {
		return nil, err
	}
	if !loginInfo.Ok() {
		return nil, loginInfo.Err()
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
	defer func() { _ = resp.Body.Close() }()
	if err := scanJson(resp.Body, &webInitResponse); err != nil {
		return nil, err
	}
	if !webInitResponse.BaseResponse.Ok() {
		return nil, webInitResponse.BaseResponse.Err()
	}
	return &webInitResponse, nil
}

// WebWxStatusNotify 通知手机已登录
func (c *Caller) WebWxStatusNotify(request *BaseRequest, response *WebInitResponse, info *LoginInfo) error {
	resp, err := c.Client.WebWxStatusNotify(request, response, info)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// SyncCheck 异步获取是否有新的消息
func (c *Caller) SyncCheck(request *BaseRequest, info *LoginInfo, response *WebInitResponse) (*SyncCheckResponse, error) {
	resp, err := c.Client.SyncCheck(request, info, response)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	results := syncCheckRegexp.FindSubmatch(buffer.Bytes())
	if len(results) != 3 {
		return nil, errors.New("parse sync key failed")
	}
	retCode, selector := string(results[1]), Selector(results[2])
	syncCheckResponse := &SyncCheckResponse{RetCode: retCode, Selector: selector}
	return syncCheckResponse, nil
}

// WebWxGetContact 获取所有的联系人
func (c *Caller) WebWxGetContact(info *LoginInfo) (Members, error) {
	var members Members
	var reqs int64
	for {
		resp, err := c.Client.WebWxGetContact(info, reqs)
		if err != nil {
			return nil, err
		}
		var item WebWxContactResponse
		if err = scanJson(resp.Body, &item); err != nil {
			_ = resp.Body.Close()
			return nil, err
		}
		if err = resp.Body.Close(); err != nil {
			return nil, err
		}
		if !item.BaseResponse.Ok() {
			return nil, item.BaseResponse.Err()
		}
		members = append(members, item.MemberList...)

		if item.Seq == 0 || item.Seq == reqs {
			break
		}
		reqs = item.Seq
	}
	return members, nil
}

// WebWxBatchGetContact 获取联系人的详情
// 注: Members参数的长度不要大于50
func (c *Caller) WebWxBatchGetContact(members Members, request *BaseRequest) (Members, error) {
	resp, err := c.Client.WebWxBatchGetContact(members, request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var item WebWxBatchContactResponse
	if err := scanJson(resp.Body, &item); err != nil {
		return nil, err
	}
	if !item.BaseResponse.Ok() {
		return nil, item.BaseResponse.Err()
	}
	return item.ContactList, nil
}

// WebWxSync 获取新的消息接口
func (c *Caller) WebWxSync(request *BaseRequest, response *WebInitResponse, info *LoginInfo) (*WebWxSyncResponse, error) {
	resp, err := c.Client.WebWxSync(request, response, info)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var webWxSyncResponse WebWxSyncResponse
	if err := scanJson(resp.Body, &webWxSyncResponse); err != nil {
		return nil, err
	}
	return &webWxSyncResponse, nil
}

// WebWxSendMsg 发送消息接口
func (c *Caller) WebWxSendMsg(msg *SendMessage, info *LoginInfo, request *BaseRequest) (*SentMessage, error) {
	resp, err := c.Client.WebWxSendMsg(msg, info, request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(msg)
}

// WebWxOplog 修改用户备注接口
func (c *Caller) WebWxOplog(request *BaseRequest, remarkName, toUserName string) error {
	resp, err := c.Client.WebWxOplog(request, remarkName, toUserName)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

func (c *Caller) UploadMedia(file *os.File, request *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*UploadResponse, error) {
	// 首先尝试上传图片
	resp, err := c.Client.WebWxUploadMediaByChunk(file, request, info, fromUserName, toUserName)
	// 无错误上传成功之后获取请求结果，判断结果是否正常
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var item UploadResponse

	if err := scanJson(resp.Body, &item); err != nil {
		return &item, err
	}
	if !item.BaseResponse.Ok() {
		return &item, item.BaseResponse.Err()
	}
	if len(item.MediaId) == 0 {
		return &item, errors.New("upload failed")
	}
	return &item, nil
}

// WebWxSendImageMsg 发送图片消息接口
func (c *Caller) WebWxSendImageMsg(reader io.Reader, request *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*SentMessage, error) {
	file, cb, err := readerToFile(reader)
	if err != nil {
		return nil, err
	}
	defer cb()
	// 首先尝试上传图片
	var mediaId string
	{
		resp, err := c.UploadMedia(file, request, info, fromUserName, toUserName)
		if err != nil {
			return nil, err
		}
		mediaId = resp.MediaId
	}
	// 构造新的图片类型的信息
	msg := NewMediaSendMessage(MsgTypeImage, fromUserName, toUserName, mediaId)
	// 发送图片信息
	resp, err := c.Client.WebWxSendMsgImg(msg, request, info)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(msg)
}

func (c *Caller) WebWxSendFile(reader io.Reader, req *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*SentMessage, error) {
	file, cb, err := readerToFile(reader)
	if err != nil {
		return nil, err
	}
	defer cb()
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

func (c *Caller) WebWxSendVideoMsg(reader io.Reader, request *BaseRequest, info *LoginInfo, fromUserName, toUserName string) (*SentMessage, error) {
	file, cb, err := readerToFile(reader)
	if err != nil {
		return nil, err
	}
	defer cb()
	var mediaId string
	{
		resp, err := c.UploadMedia(file, request, info, fromUserName, toUserName)
		if err != nil {
			return nil, err
		}
		mediaId = resp.MediaId
	}
	// 构造新的图片类型的信息
	msg := NewMediaSendMessage(MsgTypeVideo, fromUserName, toUserName, mediaId)
	resp, err := c.Client.WebWxSendVideoMsg(request, msg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(msg)
}

// WebWxSendAppMsg 发送媒体消息
func (c *Caller) WebWxSendAppMsg(msg *SendMessage, req *BaseRequest) (*SentMessage, error) {
	resp, err := c.Client.WebWxSendAppMsg(msg, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(msg)
}

// Logout 用户退出
func (c *Caller) Logout(info *LoginInfo) error {
	resp, err := c.Client.Logout(info)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
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
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
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
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// WebWxVerifyUser 同意加好友请求
func (c *Caller) WebWxVerifyUser(storage *Storage, info RecommendInfo, verifyContent string) error {
	resp, err := c.Client.WebWxVerifyUser(storage, info, verifyContent)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// WebWxRevokeMsg 撤回消息操作
func (c *Caller) WebWxRevokeMsg(msg *SentMessage, request *BaseRequest) error {
	resp, err := c.Client.WebWxRevokeMsg(msg, request)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// WebWxStatusAsRead 将消息设置为已读
func (c *Caller) WebWxStatusAsRead(request *BaseRequest, info *LoginInfo, msg *Message) error {
	resp, err := c.Client.WebWxStatusAsRead(request, info, msg)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// WebWxRelationPin 将联系人是否置顶
func (c *Caller) WebWxRelationPin(request *BaseRequest, user *User, op uint8) error {
	resp, err := c.Client.WebWxRelationPin(request, op, user)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// WebWxPushLogin 免扫码登陆接口
func (c *Caller) WebWxPushLogin(uin int64) (*PushLoginResponse, error) {
	resp, err := c.Client.WebWxPushLogin(uin)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var item PushLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

// WebWxCreateChatRoom 创建群聊
func (c *Caller) WebWxCreateChatRoom(request *BaseRequest, info *LoginInfo, topic string, friends Friends) (*Group, error) {
	resp, err := c.Client.WebWxCreateChatRoom(request, info, topic, friends)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var item struct {
		BaseResponse BaseResponse
		ChatRoomName string
	}
	if err = json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	if !item.BaseResponse.Ok() {
		return nil, item.BaseResponse.Err()
	}
	group := Group{User: &User{UserName: item.ChatRoomName}}
	return &group, nil
}

// WebWxRenameChatRoom 群组重命名
func (c *Caller) WebWxRenameChatRoom(request *BaseRequest, info *LoginInfo, newTopic string, group *Group) error {
	resp, err := c.Client.WebWxRenameChatRoom(request, info, newTopic, group)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// SetMode 设置Client的模式
func (c *Client) SetMode(mode Mode) {
	c.mode = mode
}

// MessageResponseParser 消息响应解析器
type MessageResponseParser struct {
	Reader io.Reader
}

// Err 解析错误
func (p *MessageResponseParser) Err() error {
	var item struct{ BaseResponse BaseResponse }
	if err := scanJson(p.Reader, &item); err != nil {
		return err
	}
	if !item.BaseResponse.Ok() {
		return item.BaseResponse.Err()
	}
	return nil
}

// MsgID 解析消息ID
func (p *MessageResponseParser) MsgID() (string, error) {
	var messageResp MessageResponse
	if err := scanJson(p.Reader, &messageResp); err != nil {
		return "", err
	}
	if !messageResp.BaseResponse.Ok() {
		return "", messageResp.BaseResponse.Err()
	}
	return messageResp.MsgID, nil
}

// SentMessage 返回 SentMessage
func (p *MessageResponseParser) SentMessage(msg *SendMessage) (*SentMessage, error) {
	msgID, err := p.MsgID()
	if err != nil {
		return nil, err
	}
	return &SentMessage{MsgId: msgID, SendMessage: msg}, nil
}

func readerToFile(reader io.Reader) (file *os.File, cb func(), err error) {
	var ok bool
	if file, ok = reader.(*os.File); ok {
		return file, func() {}, nil
	}
	file, err = os.CreateTemp("", "*")
	if err != nil {
		return nil, nil, err
	}
	cb = func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}
	_, err = io.Copy(file, reader)
	if err != nil {
		cb()
		return nil, nil, err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		cb()
		return nil, nil, err
	}
	return file, cb, nil
}
