package openwechat

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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
func (c *Caller) GetLoginUUID(ctx context.Context) (string, error) {
	resp, err := c.Client.GetLoginUUID(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var buffer bytes.Buffer
	if _, err = buffer.ReadFrom(resp.Body); err != nil {
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
func (c *Caller) CheckLogin(ctx context.Context, uuid, tip string) (CheckLoginResponse, error) {
	resp, err := c.Client.CheckLogin(ctx, uuid, tip)
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
func (c *Caller) GetLoginInfo(ctx context.Context, path *url.URL) (*LoginInfo, error) {
	// 从响应体里面获取需要跳转的url
	query := path.Query()
	query.Set("version", "v2")
	path.RawQuery = query.Encode()
	resp, err := c.Client.GetLoginInfo(ctx, path)
	if err != nil {
		return nil, err
	}
	// 微信 v2 版本修复了301 response missing Location header 的问题
	defer func() { _ = resp.Body.Close() }()

	// 这里部分账号可能会被误判, 但是我又没有号测试。如果你遇到了这个问题，可以帮忙解决一下。😊
	if _, exists := CookieGroup(resp.Cookies()).GetByName("wxuin"); !exists {
		err = ErrForbidden
		if c.Client.mode != desktop {
			err = fmt.Errorf("%w: try to login with desktop mode", err)
		}
		return nil, err
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var loginInfo LoginInfo

	// xml结构体序列化储存
	// 为什么这里不直接使用resp.Body?
	// 因为要确保传入的reader实现了 io.ByteReader 接口
	// https://github.com/eatmoreapple/openwechat/pull/345
	if err = xml.NewDecoder(bytes.NewBuffer(bs)).Decode(&loginInfo); err != nil {
		return nil, err
	}
	if !loginInfo.Ok() {
		return nil, loginInfo.Err()
	}
	// set domain
	c.Client.Domain = WechatDomain(path.Host)
	return &loginInfo, nil
}

// WebInit 获取初始化信息
func (c *Caller) WebInit(ctx context.Context, request *BaseRequest) (*WebInitResponse, error) {
	resp, err := c.Client.WebInit(ctx, request)
	if err != nil {
		return nil, err
	}
	var webInitResponse WebInitResponse
	defer func() { _ = resp.Body.Close() }()
	if err = json.NewDecoder(resp.Body).Decode(&webInitResponse); err != nil {
		return nil, err
	}
	if !webInitResponse.BaseResponse.Ok() {
		return nil, webInitResponse.BaseResponse.Err()
	}
	return &webInitResponse, nil
}

type CallerCommonOptions struct {
	BaseRequest     *BaseRequest
	WebInitResponse *WebInitResponse
	LoginInfo       *LoginInfo
}

type CallerWebWxStatusNotifyOptions CallerCommonOptions

// WebWxStatusNotify 通知手机已登录
func (c *Caller) WebWxStatusNotify(ctx context.Context, opt *CallerWebWxStatusNotifyOptions) error {
	notifyOpt := &ClientWebWxStatusNotifyOptions{
		BaseRequest:     opt.BaseRequest,
		WebInitResponse: opt.WebInitResponse,
		LoginInfo:       opt.LoginInfo,
	}
	resp, err := c.Client.WebWxStatusNotify(ctx, notifyOpt)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

type CallerSyncCheckOptions CallerCommonOptions

// SyncCheck 异步获取是否有新的消息
func (c *Caller) SyncCheck(ctx context.Context, opt *CallerSyncCheckOptions) (*SyncCheckResponse, error) {
	syncCheckOption := &ClientSyncCheckOptions{
		BaseRequest:     opt.BaseRequest,
		WebInitResponse: opt.WebInitResponse,
		LoginInfo:       opt.LoginInfo,
	}
	resp, err := c.Client.SyncCheck(ctx, syncCheckOption)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var buffer bytes.Buffer
	if _, err = buffer.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return NewSyncCheckResponse(buffer.Bytes())
}

// WebWxGetContact 获取所有的联系人
func (c *Caller) WebWxGetContact(ctx context.Context, info *LoginInfo) (Members, error) {
	var members Members
	var reqs int64
	for {
		resp, err := c.Client.WebWxGetContact(ctx, info.SKey, reqs)
		if err != nil {
			return nil, err
		}
		var item WebWxContactResponse
		if err = json.NewDecoder(resp.Body).Decode(&item); err != nil {
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
func (c *Caller) WebWxBatchGetContact(ctx context.Context, members Members, request *BaseRequest) (Members, error) {
	resp, err := c.Client.WebWxBatchGetContact(ctx, members, request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var item WebWxBatchContactResponse
	if err = json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	if !item.BaseResponse.Ok() {
		return nil, item.BaseResponse.Err()
	}
	return item.ContactList, nil
}

type CallerWebWxSyncOptions CallerCommonOptions

// WebWxSync 获取新的消息接口
func (c *Caller) WebWxSync(ctx context.Context, opt *CallerWebWxSyncOptions) (*WebWxSyncResponse, error) {
	wxSyncOption := &ClientWebWxSyncOptions{
		BaseRequest:     opt.BaseRequest,
		WebInitResponse: opt.WebInitResponse,
		LoginInfo:       opt.LoginInfo,
	}
	resp, err := c.Client.WebWxSync(ctx, wxSyncOption)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var webWxSyncResponse WebWxSyncResponse
	if err = json.NewDecoder(resp.Body).Decode(&webWxSyncResponse); err != nil {
		return nil, err
	}
	return &webWxSyncResponse, nil
}

type CallerWebWxSendMsgOptions struct {
	LoginInfo   *LoginInfo
	BaseRequest *BaseRequest
	Message     *SendMessage
}

// WebWxSendMsg 发送消息接口
func (c *Caller) WebWxSendMsg(ctx context.Context, opt *CallerWebWxSendMsgOptions) (*SentMessage, error) {
	wxSendMsgOption := &ClientWebWxSendMsgOptions{
		BaseRequest: opt.BaseRequest,
		LoginInfo:   opt.LoginInfo,
		Message:     opt.Message,
	}
	resp, err := c.Client.WebWxSendMsg(ctx, wxSendMsgOption)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(opt.Message)
}

type CallerWebWxOplogOptions struct {
	RemarkName  string
	ToUserName  string
	BaseRequest *BaseRequest
}

// WebWxOplog 修改用户备注接口
func (c *Caller) WebWxOplog(ctx context.Context, opt *CallerWebWxOplogOptions) error {
	wxOpLogOption := &ClientWebWxOplogOption{
		RemarkName:  opt.RemarkName,
		UserName:    opt.ToUserName,
		BaseRequest: opt.BaseRequest,
	}
	resp, err := c.Client.WebWxOplog(ctx, wxOpLogOption)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

type CallerUploadMediaOptions struct {
	FromUserName string
	ToUserName   string
	File         *os.File
	BaseRequest  *BaseRequest
	LoginInfo    *LoginInfo
}

func (c *Caller) UploadMedia(ctx context.Context, opt *CallerUploadMediaOptions) (*UploadResponse, error) {
	// 首先尝试上传图片
	clientWebWxUploadMediaByChunkOpt := &ClientWebWxUploadMediaByChunkOptions{
		FromUserName: opt.FromUserName,
		ToUserName:   opt.ToUserName,
		File:         opt.File,
		BaseRequest:  opt.BaseRequest,
		LoginInfo:    opt.LoginInfo,
	}
	resp, err := c.Client.WebWxUploadMediaByChunk(ctx, clientWebWxUploadMediaByChunkOpt)
	// 无错误上传成功之后获取请求结果，判断结果是否正常
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var item UploadResponse
	if err = json.NewDecoder(resp.Body).Decode(&item); err != nil {
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

type CallerUploadMediaCommonOptions struct {
	FromUserName string
	ToUserName   string
	Reader       io.Reader
	BaseRequest  *BaseRequest
	LoginInfo    *LoginInfo
}

type CallerWebWxSendImageMsgOptions CallerUploadMediaCommonOptions

// WebWxSendImageMsg 发送图片消息接口
func (c *Caller) WebWxSendImageMsg(ctx context.Context, opt *CallerWebWxSendImageMsgOptions) (*SentMessage, error) {
	file, cb, err := readerToFile(opt.Reader)
	if err != nil {
		return nil, err
	}
	defer cb()
	// 首先尝试上传图片
	var mediaId string
	{
		uploadMediaOption := &CallerUploadMediaOptions{
			FromUserName: opt.FromUserName,
			ToUserName:   opt.ToUserName,
			File:         file,
			BaseRequest:  opt.BaseRequest,
			LoginInfo:    opt.LoginInfo,
		}
		resp, err := c.UploadMedia(ctx, uploadMediaOption)
		if err != nil {
			return nil, err
		}
		mediaId = resp.MediaId
	}
	// 构造新的图片类型的信息
	msg := NewMediaSendMessage(MsgTypeImage, opt.FromUserName, opt.ToUserName, mediaId)
	// 发送图片信息
	sendImageOption := &ClientWebWxSendMsgOptions{
		BaseRequest: opt.BaseRequest,
		LoginInfo:   opt.LoginInfo,
		Message:     msg,
	}
	resp, err := c.Client.WebWxSendMsgImg(ctx, sendImageOption)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(msg)
}

type CallerWebWxSendFileOptions CallerUploadMediaCommonOptions

func (c *Caller) WebWxSendFile(ctx context.Context, opt *CallerWebWxSendFileOptions) (*SentMessage, error) {
	file, cb, err := readerToFile(opt.Reader)
	if err != nil {
		return nil, err
	}
	defer cb()

	uploadMediaOption := &CallerUploadMediaOptions{
		FromUserName: opt.FromUserName,
		ToUserName:   opt.ToUserName,
		File:         file,
		BaseRequest:  opt.BaseRequest,
		LoginInfo:    opt.LoginInfo,
	}
	resp, err := c.UploadMedia(ctx, uploadMediaOption)
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
	msg := NewSendMessage(AppMessage, string(content), opt.FromUserName, opt.ToUserName, "")
	return c.WebWxSendAppMsg(ctx, msg, opt.BaseRequest)
}

type CallerWebWxSendAppMsgOptions CallerUploadMediaCommonOptions

func (c *Caller) WebWxSendVideoMsg(ctx context.Context, opt *CallerWebWxSendAppMsgOptions) (*SentMessage, error) {
	file, cb, err := readerToFile(opt.Reader)
	if err != nil {
		return nil, err
	}
	defer cb()
	var mediaId string
	{
		uploadMediaOption := &CallerUploadMediaOptions{
			FromUserName: opt.FromUserName,
			ToUserName:   opt.ToUserName,
			File:         file,
			BaseRequest:  opt.BaseRequest,
			LoginInfo:    opt.LoginInfo,
		}

		resp, err := c.UploadMedia(ctx, uploadMediaOption)
		if err != nil {
			return nil, err
		}
		mediaId = resp.MediaId
	}
	// 构造新的图片类型的信息
	msg := NewMediaSendMessage(MsgTypeVideo, opt.FromUserName, opt.ToUserName, mediaId)
	resp, err := c.Client.WebWxSendVideoMsg(ctx, opt.BaseRequest, msg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(msg)
}

// WebWxSendAppMsg 发送媒体消息
func (c *Caller) WebWxSendAppMsg(ctx context.Context, msg *SendMessage, req *BaseRequest) (*SentMessage, error) {
	resp, err := c.Client.WebWxSendAppMsg(ctx, msg, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.SentMessage(msg)
}

// Logout 用户退出
func (c *Caller) Logout(ctx context.Context, info *LoginInfo) error {
	resp, err := c.Client.Logout(ctx, info)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

type CallerAddFriendIntoChatRoomOptions struct {
	BaseRequest *BaseRequest
	LoginInfo   *LoginInfo
	Group       *Group
	Friends     []*Friend
}

// AddFriendIntoChatRoom 拉好友入群
func (c *Caller) AddFriendIntoChatRoom(ctx context.Context, opt *CallerAddFriendIntoChatRoomOptions) error {
	if len(opt.Friends) == 0 {
		return errors.New("no friends found")
	}
	inviteMemberList := make([]string, len(opt.Friends))
	for i, friend := range opt.Friends {
		inviteMemberList[i] = friend.UserName
	}
	clientAddMemberIntoChatRoomOption := &ClientAddMemberIntoChatRoomOption{
		BaseRequest:      opt.BaseRequest,
		LoginInfo:        opt.LoginInfo,
		Group:            opt.Group.UserName,
		InviteMemberList: inviteMemberList,
	}
	resp, err := c.Client.AddMemberIntoChatRoom(ctx, clientAddMemberIntoChatRoomOption)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

type CallerRemoveFriendFromChatRoomOptions struct {
	BaseRequest *BaseRequest
	LoginInfo   *LoginInfo
	Group       *Group
	Members     []*User
}

// RemoveFriendFromChatRoom 从群聊中移除用户
func (c *Caller) RemoveFriendFromChatRoom(ctx context.Context, opt *CallerRemoveFriendFromChatRoomOptions) error {
	if len(opt.Members) == 0 {
		return errors.New("no users found")
	}
	users := make([]string, len(opt.Members))
	for i, member := range opt.Members {
		users[i] = member.UserName
	}
	req := &ClientRemoveMemberFromChatRoomOption{
		BaseRequest:   opt.BaseRequest,
		LoginInfo:     opt.LoginInfo,
		Group:         opt.Group.UserName,
		DelMemberList: users,
	}
	resp, err := c.Client.RemoveMemberFromChatRoom(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

type CallerWebWxVerifyUserOptions struct {
	VerifyContent string
	RecommendInfo RecommendInfo
	BaseRequest   *BaseRequest
	LoginInfo     *LoginInfo
}

// WebWxVerifyUser 同意加好友请求
func (c *Caller) WebWxVerifyUser(ctx context.Context, opt *CallerWebWxVerifyUserOptions) error {
	webWxVerifyUserOption := &ClientWebWxVerifyUserOption{
		BaseRequest:   opt.BaseRequest,
		LoginInfo:     opt.LoginInfo,
		VerifyContent: opt.VerifyContent,
		RecommendInfo: opt.RecommendInfo,
	}
	resp, err := c.Client.WebWxVerifyUser(ctx, webWxVerifyUserOption)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// WebWxRevokeMsg 撤回消息操作
func (c *Caller) WebWxRevokeMsg(ctx context.Context, msg *SentMessage, request *BaseRequest) error {
	resp, err := c.Client.WebWxRevokeMsg(ctx, msg, request)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

type CallerWebWxStatusAsReadOptions struct {
	BaseRequest *BaseRequest
	LoginInfo   *LoginInfo
	Message     *Message
}

// WebWxStatusAsRead 将消息设置为已读
func (c *Caller) WebWxStatusAsRead(ctx context.Context, opt *CallerWebWxStatusAsReadOptions) error {
	statusAsReadOption := &ClientWebWxStatusAsReadOption{
		Request:   opt.BaseRequest,
		LoginInfo: opt.LoginInfo,
		Message:   opt.Message,
	}
	resp, err := c.Client.WebWxStatusAsRead(ctx, statusAsReadOption)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

type CallerWebWxRelationPinOptions struct {
	BaseRequest *BaseRequest
	User        *User
	Op          uint8
}

// WebWxRelationPin 将联系人是否置顶
func (c *Caller) WebWxRelationPin(ctx context.Context, opt *CallerWebWxRelationPinOptions) error {
	webWxRelationPinOption := &ClientWebWxRelationPinOption{
		Request:    opt.BaseRequest,
		Op:         opt.Op,
		RemarkName: opt.User.RemarkName,
		UserName:   opt.User.UserName,
	}
	resp, err := c.Client.WebWxRelationPin(ctx, webWxRelationPinOption)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	parser := MessageResponseParser{resp.Body}
	return parser.Err()
}

// WebWxPushLogin 免扫码登陆接口
func (c *Caller) WebWxPushLogin(ctx context.Context, uin int64) (*PushLoginResponse, error) {
	resp, err := c.Client.WebWxPushLogin(ctx, uin)
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

type CallerWebWxCreateChatRoomOptions struct {
	BaseRequest *BaseRequest
	LoginInfo   *LoginInfo
	Topic       string
	Friends     Friends
}

// WebWxCreateChatRoom 创建群聊
func (c *Caller) WebWxCreateChatRoom(ctx context.Context, opt *CallerWebWxCreateChatRoomOptions) (*Group, error) {
	if len(opt.Friends) == 0 {
		return nil, errors.New("create group with no friends")
	}
	friends := make([]string, len(opt.Friends))
	for i, friend := range opt.Friends {
		friends[i] = friend.UserName
	}
	webWxCreateChatRoomOption := &ClientWebWxCreateChatRoomOption{
		Request:   opt.BaseRequest,
		Topic:     opt.Topic,
		Friends:   friends,
		LoginInfo: opt.LoginInfo,
	}
	resp, err := c.Client.WebWxCreateChatRoom(ctx, webWxCreateChatRoomOption)
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

type CallerWebWxRenameChatRoomOptions struct {
	NewTopic    string
	BaseRequest *BaseRequest
	LoginInfo   *LoginInfo
	Group       *Group
}

// WebWxRenameChatRoom 群组重命名
func (c *Caller) WebWxRenameChatRoom(ctx context.Context, opt *CallerWebWxRenameChatRoomOptions) error {
	webWxRenameChatRoomOption := &ClientWebWxRenameChatRoomOption{
		Request:   opt.BaseRequest,
		NewTopic:  opt.NewTopic,
		Group:     opt.Group.UserName,
		LoginInfo: opt.LoginInfo,
	}
	resp, err := c.Client.WebWxRenameChatRoom(ctx, webWxRenameChatRoomOption)
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
	if err := json.NewDecoder(p.Reader).Decode(&item); err != nil {
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
	if err := json.NewDecoder(p.Reader).Decode(&messageResp); err != nil {
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
