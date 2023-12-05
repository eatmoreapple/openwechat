package openwechat

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// User 抽象的用户结构: 好友 群组 公众号
type User struct {
	HideInputBarFlag  int
	StarFriend        int
	Sex               int
	AppAccountFlag    int
	VerifyFlag        int
	ContactFlag       int
	WebWxPluginSwitch int
	HeadImgFlag       int
	SnsFlag           int
	IsOwner           int
	MemberCount       int
	ChatRoomId        int
	UniFriend         int
	OwnerUin          int
	Statues           int
	AttrStatus        int64
	Uin               int64
	Province          string
	City              string
	Alias             string
	DisplayName       string
	KeyWord           string
	EncryChatRoomId   string
	UserName          string
	NickName          string
	HeadImgUrl        string
	RemarkName        string
	PYInitial         string
	PYQuanPin         string
	RemarkPYInitial   string
	RemarkPYQuanPin   string
	Signature         string

	MemberList Members

	self *Self
}

// implement fmt.Stringer
func (u *User) String() string {
	format := "User"
	if u.IsSelf() {
		format = "Self"
	} else if u.IsFriend() {
		format = "Friend"
	} else if u.IsGroup() {
		format = "Group"
	} else if u.IsMP() {
		format = "MP"
	}
	return fmt.Sprintf("<%s:%s>", format, u.NickName)
}

// GetAvatarResponse 获取用户头像
func (u *User) GetAvatarResponse() (resp *http.Response, err error) {
	for i := 0; i < 3; i++ {
		resp, err = u.self.bot.Caller.Client.WebWxGetHeadImg(u.Self().Bot().Context(), u)
		if err != nil {
			return nil, err
		}
		// 这里存在 ContentLength 为0的情况，需要重试
		if resp.ContentLength > 0 {
			break
		}
	}
	return resp, err
}

// SaveAvatar 下载用户头像
func (u *User) SaveAvatar(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return u.SaveAvatarWithWriter(file)
}

func (u *User) SaveAvatarWithWriter(writer io.Writer) error {
	resp, err := u.GetAvatarResponse()
	if err != nil {
		return err
	}
	// 写文件前判断下 content length 是否是 0，不然保存的头像会出现
	// image not loaded  try to open it externally to fix format problem 问题
	if resp.ContentLength == 0 {
		return errors.New("get avatar response content length is 0")
	}
	defer func() { _ = resp.Body.Close() }()
	_, err = io.Copy(writer, resp.Body)
	return err
}

// Detail 获取用户的详情
func (u *User) Detail() error {
	if u.UserName == u.self.UserName {
		return nil
	}
	members := Members{u}
	if err := members.Detail(); err != nil {
		return err
	}
	*u = *members.First()
	u.MemberList.init(u.self)
	return nil
}

// IsFriend 判断是否为好友
func (u *User) IsFriend() bool {
	return !u.IsGroup() && strings.HasPrefix(u.UserName, "@") && u.VerifyFlag == 0
}

// AsFriend 将当前用户转换为好友类型
func (u *User) AsFriend() (*Friend, bool) {
	if u.IsFriend() {
		return &Friend{User: u}, true
	}
	return nil, false
}

// IsGroup 判断是否为群组
func (u *User) IsGroup() bool {
	return strings.HasPrefix(u.UserName, "@@") && u.VerifyFlag == 0
}

// AsGroup 将当前用户转换为群组类型
func (u *User) AsGroup() (*Group, bool) {
	if u.IsGroup() {
		return &Group{User: u}, true
	}
	return nil, false
}

// IsMP  判断是否为公众号
func (u *User) IsMP() bool {
	return u.VerifyFlag == 8 || u.VerifyFlag == 24 || u.VerifyFlag == 136
}

// AsMP 将当前用户转换为公众号类型
func (u *User) AsMP() (*Mp, bool) {
	if u.IsMP() {
		return &Mp{User: u}, true
	}
	return nil, false
}

// Pin 将联系人置顶
func (u *User) Pin() error {
	opt := &CallerWebWxRelationPinOptions{
		BaseRequest: u.self.bot.Storage.Request,
		User:        u,
		Op:          1,
	}
	return u.self.bot.Caller.WebWxRelationPin(u.Self().Bot().Context(), opt)
}

// UnPin 将联系人取消置顶
func (u *User) UnPin() error {
	opt := &CallerWebWxRelationPinOptions{
		BaseRequest: u.self.bot.Storage.Request,
		User:        u,
		Op:          0,
	}
	return u.self.bot.Caller.WebWxRelationPin(u.Self().Bot().Context(), opt)
}

// IsPin 判断当前联系人(好友、群组、公众号)是否为置顶状态
func (u *User) IsPin() bool {
	return u.ContactFlag == 2051
}

// ID 获取用户头像id
// Deprecated: 请使用 AvatarID
func (u *User) ID() string {
	return u.AvatarID()
}

// AvatarID 获取用户头像id
// 这个值会随着用户更换头像而变化
func (u *User) AvatarID() string {
	// 首先尝试获取uid
	if u.Uin != 0 {
		return strconv.FormatInt(u.Uin, 10)
	}
	// 如果uid不存在，尝试从头像url中获取
	if u.HeadImgUrl != "" {
		index := strings.Index(u.HeadImgUrl, "?") + 1
		if len(u.HeadImgUrl) > index {
			query := u.HeadImgUrl[index:]
			params, err := url.ParseQuery(query)
			if err != nil {
				return ""
			}
			return params.Get("seq")

		}
	}
	return ""
}

// Equal 判断两个用户是否相等
func (u *User) Equal(user *User) bool {
	// invalid user is not equal to any user
	if u == nil || user == nil {
		return false
	}
	// not came from same bot
	if u.Self() != user.Self() {
		return false
	}
	return u.UserName == user.UserName
}

// Self 返回当前用户
func (u *User) Self() *Self {
	return u.self
}

// IsSelf 判断是否为当前用户
func (u *User) IsSelf() bool {
	return u.UserName == u.Self().UserName
}

// OrderSymbol 获取用户的排序标识
func (u *User) OrderSymbol() string {
	var symbol string
	if u.RemarkPYQuanPin != "" {
		symbol = u.RemarkPYQuanPin
	} else if u.PYQuanPin != "" {
		symbol = u.PYQuanPin
	} else {
		symbol = u.NickName
	}
	symbol = html.UnescapeString(symbol)
	symbol = strings.ToUpper(symbol)
	symbol = regexp.MustCompile("/\\W/ig").ReplaceAllString(symbol, "")
	if len(symbol) > 0 && symbol[0] < 'A' {
		return "~"
	}
	return symbol
}

// 格式化emoji表情
func (u *User) formatEmoji() {
	u.NickName = FormatEmoji(u.NickName)
	u.RemarkName = FormatEmoji(u.RemarkName)
	u.DisplayName = FormatEmoji(u.DisplayName)
}

func newUser(self *Self, username string) *User {
	return &User{
		UserName: username,
		self:     self,
	}
}

// Self 自己,当前登录用户对象
type Self struct {
	*User
	bot        *Bot
	fileHelper *Friend
	members    Members
	friends    Friends
	groups     Groups
	mps        Mps
}

// Members 获取所有的好友、群组、公众号信息
func (s *Self) Members(update ...bool) (Members, error) {
	// 首先判断缓存里有没有,如果没有则去更新缓存
	// 判断是否需要更新,如果传入的参数不为nil,则取第一个
	if s.members == nil || (len(update) > 0 && update[0]) {
		if err := s.updateMembers(); err != nil {
			return nil, err
		}
	}
	s.members.Sort()
	return s.members, nil
}

// 更新联系人处理
func (s *Self) updateMembers() error {
	info := s.bot.Storage.LoginInfo
	members, err := s.bot.Caller.WebWxGetContact(s.Bot().Context(), info)
	if err != nil {
		return err
	}
	members.init(s)
	s.members = members
	return nil
}

// FileHelper 获取文件传输助手对象，封装成Friend返回
//
//	fh := self.FileHelper() // or fh := openwechat.NewFriendHelper(self)
func (s *Self) FileHelper() *Friend {
	if s.fileHelper == nil {
		s.fileHelper = NewFriendHelper(s)
	}
	return s.fileHelper
}
func (s *Self) ChkFrdGrpMpNil() bool {
	return s.friends == nil && s.groups == nil && s.mps == nil
}

// Friends 获取所有的好友
func (s *Self) Friends(update ...bool) (Friends, error) {
	if (len(update) > 0 && update[0]) || s.ChkFrdGrpMpNil() {
		if _, err := s.Members(true); err != nil {
			return nil, err
		}
	}
	if s.friends == nil || (len(update) > 0 && update[0]) {
		s.friends = s.members.Friends()
	}
	return s.friends, nil
}

// Groups 获取所有的群组
func (s *Self) Groups(update ...bool) (Groups, error) {

	if (len(update) > 0 && update[0]) || s.ChkFrdGrpMpNil() {
		if _, err := s.Members(true); err != nil {
			return nil, err
		}

	}
	if s.groups == nil || (len(update) > 0 && update[0]) {
		s.groups = s.members.Groups()
	}
	return s.groups, nil
}

// Mps 获取所有的公众号
func (s *Self) Mps(update ...bool) (Mps, error) {
	if (len(update) > 0 && update[0]) || s.ChkFrdGrpMpNil() {
		if _, err := s.Members(true); err != nil {
			return nil, err
		}
	}
	if s.mps == nil || (len(update) > 0 && update[0]) {
		s.mps = s.members.MPs()
	}
	return s.mps, nil
}

// UpdateMembersDetail 更新所有的联系人信息
func (s *Self) UpdateMembersDetail() error {
	// 先获取所有的联系人
	members, err := s.Members()
	if err != nil {
		return err
	}
	return members.Detail()
}

func (s *Self) sendTextToUser(username, text string) (*SentMessage, error) {
	msg := NewTextSendMessage(text, s.UserName, username)
	opt := &CallerWebWxSendMsgOptions{
		LoginInfo:   s.bot.Storage.LoginInfo,
		BaseRequest: s.bot.Storage.Request,
		Message:     msg,
	}
	sentMessage, err := s.bot.Caller.WebWxSendMsg(s.Bot().Context(), opt)
	return s.sendMessageWrapper(sentMessage, err)
}

func (s *Self) sendImageToUser(username string, file io.Reader) (*SentMessage, error) {
	opt := &CallerWebWxSendImageMsgOptions{
		FromUserName: s.UserName,
		ToUserName:   username,
		Reader:       file,
		BaseRequest:  s.bot.Storage.Request,
		LoginInfo:    s.bot.Storage.LoginInfo,
	}
	sentMessage, err := s.bot.Caller.WebWxSendImageMsg(s.Bot().Context(), opt)
	return s.sendMessageWrapper(sentMessage, err)
}

func (s *Self) sendVideoToUser(username string, file io.Reader) (*SentMessage, error) {
	opt := &CallerWebWxSendAppMsgOptions{
		FromUserName: s.UserName,
		ToUserName:   username,
		Reader:       file,
		BaseRequest:  s.bot.Storage.Request,
		LoginInfo:    s.bot.Storage.LoginInfo,
	}
	sentMessage, err := s.bot.Caller.WebWxSendVideoMsg(s.Bot().Context(), opt)
	return s.sendMessageWrapper(sentMessage, err)
}

func (s *Self) sendFileToUser(username string, file io.Reader) (*SentMessage, error) {
	opt := &CallerWebWxSendFileOptions{
		FromUserName: s.UserName,
		ToUserName:   username,
		Reader:       file,
		BaseRequest:  s.bot.Storage.Request,
		LoginInfo:    s.bot.Storage.LoginInfo,
	}
	sentMessage, err := s.bot.Caller.WebWxSendFile(s.Bot().Context(), opt)
	return s.sendMessageWrapper(sentMessage, err)
}

// SendTextToFriend 发送文本消息给好友
func (s *Self) SendTextToFriend(friend *Friend, text string) (*SentMessage, error) {
	return s.sendTextToUser(friend.User.UserName, text)
}

// SendImageToFriend 发送图片消息给好友
func (s *Self) SendImageToFriend(friend *Friend, file io.Reader) (*SentMessage, error) {
	return s.sendImageToUser(friend.User.UserName, file)
}

// SendVideoToFriend 发送视频给好友
func (s *Self) SendVideoToFriend(friend *Friend, file io.Reader) (*SentMessage, error) {
	return s.sendVideoToUser(friend.User.UserName, file)
}

// SendFileToFriend 发送文件给好友
func (s *Self) SendFileToFriend(friend *Friend, file io.Reader) (*SentMessage, error) {
	return s.sendFileToUser(friend.User.UserName, file)
}

// SetRemarkNameToFriend 设置好友备注
// Deprecated
// 已经失效了
//
//	self.SetRemarkNameToFriend(friend, "remark") // or friend.SetRemarkName("remark")
func (s *Self) SetRemarkNameToFriend(friend *Friend, remarkName string) error {
	opt := &CallerWebWxOplogOptions{
		BaseRequest: s.bot.Storage.Request,
		ToUserName:  friend.UserName,
		RemarkName:  remarkName,
	}
	err := s.bot.Caller.WebWxOplog(s.Bot().Context(), opt)
	if err == nil {
		friend.RemarkName = remarkName
	}
	return err
}

// CreateGroup 创建群聊
// topic 群昵称,可以传递字符串
// friends 群员,最少为2个，加上自己3个,三人才能成群
func (s *Self) CreateGroup(topic string, friends ...*Friend) (*Group, error) {
	friends = Friends(friends).Uniq()
	if len(friends) < 2 {
		return nil, errors.New("a group must be at least 2 members")
	}
	opt := &CallerWebWxCreateChatRoomOptions{
		BaseRequest: s.bot.Storage.Request,
		LoginInfo:   s.bot.Storage.LoginInfo,
		Topic:       topic,
		Friends:     friends,
	}
	group, err := s.bot.Caller.WebWxCreateChatRoom(s.Bot().Context(), opt)
	if err != nil {
		return nil, err
	}
	group.self = s
	if err = group.Detail(); err != nil {
		return nil, err
	}
	// 添加到群组列表
	s.groups = append(s.groups, group)
	return group, nil
}

// AddFriendsIntoGroup 拉多名好友进群
// 最好自己是群主,成功率高一点,因为有的群允许非群组拉人,而有的群不允许
func (s *Self) AddFriendsIntoGroup(group *Group, friends ...*Friend) error {
	if len(friends) == 0 {
		return nil
	}
	friends = Friends(friends).Uniq()
	// 获取群的所有的群员
	groupMembers, err := group.Members()
	if err != nil {
		return err
	}
	// 判断当前的成员在不在群里面
	for _, friend := range friends {
		for _, member := range groupMembers {
			if member.UserName == friend.UserName {
				return fmt.Errorf("user %s has alreay in this group", friend.String())
			}
		}
	}
	opt := &CallerAddFriendIntoChatRoomOptions{
		BaseRequest: s.bot.Storage.Request,
		LoginInfo:   s.bot.Storage.LoginInfo,
		Group:       group,
		GroupLength: groupMembers.Count(),
		Friends:     friends,
	}
	return s.bot.Caller.AddFriendIntoChatRoom(s.Bot().Context(), opt)
}

// RemoveMemberFromGroup 从群聊中移除用户
// Deprecated
// 无论是网页版，还是程序上都不起作用
func (s *Self) RemoveMemberFromGroup(group *Group, members Members) error {
	if len(members) == 0 {
		return nil
	}
	if group.IsOwner == 0 {
		return errors.New("group owner required")
	}
	groupMembers, err := group.Members()
	if err != nil {
		return err
	}
	// 判断用户是否在群聊中
	var count int
	for _, member := range members {
		for _, gm := range groupMembers {
			if gm.UserName == member.UserName {
				count++
			}
		}
	}
	if count != len(members) {
		return errors.New("invalid members")
	}
	opt := &CallerRemoveFriendFromChatRoomOptions{
		BaseRequest: s.bot.Storage.Request,
		LoginInfo:   s.bot.Storage.LoginInfo,
		Group:       group,
		Members:     members,
	}
	return s.bot.Caller.RemoveFriendFromChatRoom(s.Bot().Context(), opt)
}

// AddFriendIntoManyGroups 拉好友进多个群聊
// AddFriendIntoGroups, 名字和上面的有点像
func (s *Self) AddFriendIntoManyGroups(friend *Friend, groups ...*Group) error {
	groups = Groups(groups).Uniq()
	for _, group := range groups {
		if err := s.AddFriendsIntoGroup(group, friend); err != nil {
			return err
		}
	}
	return nil
}

// RenameGroup 群组重命名
// Deprecated
func (s *Self) RenameGroup(group *Group, newName string) error {
	webWxRenameChatRoomOptions := &CallerWebWxRenameChatRoomOptions{
		BaseRequest: s.bot.Storage.Request,
		LoginInfo:   s.bot.Storage.LoginInfo,
		Group:       group,
		NewTopic:    newName,
	}
	err := s.bot.Caller.WebWxRenameChatRoom(s.Bot().Context(), webWxRenameChatRoomOptions)
	if err == nil {
		group.NickName = newName
	}
	return err
}

// SendTextToGroup 发送文本消息给群组
func (s *Self) SendTextToGroup(group *Group, text string) (*SentMessage, error) {
	return s.sendTextToUser(group.User.UserName, text)
}

// SendImageToGroup 发送图片消息给群组
func (s *Self) SendImageToGroup(group *Group, file io.Reader) (*SentMessage, error) {
	return s.sendImageToUser(group.User.UserName, file)
}

// SendVideoToGroup 发送视频给群组
func (s *Self) SendVideoToGroup(group *Group, file io.Reader) (*SentMessage, error) {
	return s.sendVideoToUser(group.User.UserName, file)
}

// SendFileToGroup 发送文件给群组
func (s *Self) SendFileToGroup(group *Group, file io.Reader) (*SentMessage, error) {
	return s.sendFileToUser(group.User.UserName, file)
}

// RevokeMessage 撤回消息
//
//	sentMessage, err := friend.SendText("message")
//	if err == nil {
//	    self.RevokeMessage(sentMessage) // or sentMessage.Revoke()
//	}
func (s *Self) RevokeMessage(msg *SentMessage) error {
	return s.bot.Caller.WebWxRevokeMsg(s.Bot().Context(), msg, s.bot.Storage.Request)
}

// 转发消息接口
func (s *Self) forwardMessage(msg *SentMessage, delay time.Duration, users ...*User) error {
	info := s.bot.Storage.LoginInfo
	req := s.bot.Storage.Request

	ctx := s.Bot().Context()

	var forwardFunc func() error
	switch msg.Type {
	case MsgTypeText:
		forwardFunc = func() error {
			opt := &CallerWebWxSendMsgOptions{
				LoginInfo:   info,
				BaseRequest: req,
				Message:     msg.SendMessage,
			}
			_, err := s.bot.Caller.WebWxSendMsg(ctx, opt)
			return err
		}
	case MsgTypeImage:
		forwardFunc = func() error {
			opt := &ClientWebWxSendMsgOptions{
				LoginInfo:   info,
				BaseRequest: req,
				Message:     msg.SendMessage,
			}
			_, err := s.bot.Caller.Client.WebWxSendMsgImg(ctx, opt)
			return err
		}
	case AppMessage:
		forwardFunc = func() error {
			_, err := s.bot.Caller.Client.WebWxSendAppMsg(ctx, msg.SendMessage, req)
			return err
		}
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
	var errGroup []error
	for _, user := range users {
		msg.FromUserName = s.UserName
		msg.ToUserName = user.UserName
		if err := forwardFunc(); err != nil {
			errGroup = append(errGroup, err)
		}
		time.Sleep(delay)
	}
	if len(errGroup) > 0 {
		return errors.Join(errGroup...)
	}
	return nil
}

// ForwardMessageToFriends 转发给好友
func (s *Self) ForwardMessageToFriends(msg *SentMessage, delay time.Duration, friends ...*Friend) error {
	members := Friends(friends).AsMembers()
	return s.forwardMessage(msg, delay, members...)
}

// ForwardMessageToGroups 转发给群组
func (s *Self) ForwardMessageToGroups(msg *SentMessage, delay time.Duration, groups ...*Group) error {
	members := Groups(groups).AsMembers()
	return s.forwardMessage(msg, delay, members...)
}

type SendMessageFunc func() (*SentMessage, error)

func (s *Self) sendMessageToMember(sendMessageFunc SendMessageFunc, delay time.Duration, members ...*User) error {
	if len(members) == 0 {
		return nil
	}
	msg, err := sendMessageFunc()
	if err != nil {
		return err
	}
	return s.forwardMessage(msg, delay, members...)
}

// sendTextToMembers 发送文本消息给群组或者好友
func (s *Self) sendTextToMembers(text string, delay time.Duration, members ...*User) error {
	if len(members) == 0 {
		return nil
	}
	var sendMessageFunc SendMessageFunc = func() (*SentMessage, error) {
		user := members[0]
		return s.sendTextToUser(user.UserName, text)
	}
	return s.sendMessageToMember(sendMessageFunc, delay, members[1:]...)
}

// sendImageToMembers 发送图片消息给群组或者好友
func (s *Self) sendImageToMembers(img io.Reader, delay time.Duration, members ...*User) error {
	if len(members) == 0 {
		return nil
	}
	var sendMessageFunc SendMessageFunc = func() (*SentMessage, error) {
		user := members[0]
		return s.sendImageToUser(user.UserName, img)
	}
	return s.sendMessageToMember(sendMessageFunc, delay, members[1:]...)
}

// sendVideoToMembers 发送视频消息给群组或者好友
func (s *Self) sendVideoToMembers(video io.Reader, delay time.Duration, members ...*User) error {
	if len(members) == 0 {
		return nil
	}
	var sendMessageFunc SendMessageFunc = func() (*SentMessage, error) {
		user := members[0]
		return s.sendVideoToUser(user.UserName, video)
	}
	return s.sendMessageToMember(sendMessageFunc, delay, members[1:]...)
}

// sendFileToMembers 发送文件消息给群组或者好友
func (s *Self) sendFileToMembers(file io.Reader, delay time.Duration, members ...*User) error {
	if len(members) == 0 {
		return nil
	}
	var sendMessageFunc SendMessageFunc = func() (*SentMessage, error) {
		user := members[0]
		return s.sendFileToUser(user.UserName, file)
	}
	return s.sendMessageToMember(sendMessageFunc, delay, members[1:]...)
}

// SendTextToFriends 发送文本消息给好友
func (s *Self) SendTextToFriends(text string, delay time.Duration, friends ...*Friend) error {
	members := Friends(friends).AsMembers()
	return s.sendTextToMembers(text, delay, members...)
}

// SendImageToFriends 发送图片消息给好友
func (s *Self) SendImageToFriends(img io.Reader, delay time.Duration, friends ...*Friend) error {
	members := Friends(friends).AsMembers()
	return s.sendImageToMembers(img, delay, members...)
}

// SendFileToFriends 发送文件给好友
func (s *Self) SendFileToFriends(file io.Reader, delay time.Duration, friends ...*Friend) error {
	members := Friends(friends).AsMembers()
	return s.sendFileToMembers(file, delay, members...)
}

// SendVideoToFriends 发送视频给好友
func (s *Self) SendVideoToFriends(video io.Reader, delay time.Duration, friends ...*Friend) error {
	members := Friends(friends).AsMembers()
	return s.sendVideoToMembers(video, delay, members...)
}

// SendTextToGroups 发送文本消息给群组
func (s *Self) SendTextToGroups(text string, delay time.Duration, groups ...*Group) error {
	members := Groups(groups).AsMembers()
	return s.sendTextToMembers(text, delay, members...)
}

// SendImageToGroups 发送图片消息给群组
func (s *Self) SendImageToGroups(img io.Reader, delay time.Duration, groups ...*Group) error {
	members := Groups(groups).AsMembers()
	return s.sendImageToMembers(img, delay, members...)
}

// SendFileToGroups 发送文件给群组
func (s *Self) SendFileToGroups(file io.Reader, delay time.Duration, groups ...*Group) error {
	members := Groups(groups).AsMembers()
	return s.sendFileToMembers(file, delay, members...)
}

// SendVideoToGroups 发送视频给群组
func (s *Self) SendVideoToGroups(video io.Reader, delay time.Duration, groups ...*Group) error {
	members := Groups(groups).AsMembers()
	return s.sendVideoToMembers(video, delay, members...)
}

// ContactList 获取最近的联系人列表
func (s *Self) ContactList() Members {
	return s.Bot().Storage.Response.ContactList
}

// MPSubscribeList 获取部分公众号文章列表
func (s *Self) MPSubscribeList() []*MPSubscribeMsg {
	return s.Bot().Storage.Response.MPSubscribeMsgList
}

// ID 当前登录用户的ID
func (s *Self) ID() int64 {
	return s.Uin
}

// Members 抽象的用户组
type Members []*User

// Uniq Members 去重
func (m Members) Uniq() Members {
	var uniqMembers = make(map[string]*User)
	for _, member := range m {
		uniqMembers[member.UserName] = member
	}
	var members = make(Members, 0, len(uniqMembers))
	for _, member := range uniqMembers {
		members = append(members, member)
	}
	return members
}

// Sort 对联系人进行排序
func (m Members) Sort() Members {
	sort.Slice(m, func(i, j int) bool { return m[i].OrderSymbol() < m[j].OrderSymbol() })
	return m
}

// Count 统计数量
func (m Members) Count() int {
	return len(m)
}

// First 获取第一个
func (m Members) First() *User {
	if m.Count() > 0 {
		u := m[0]
		return u
	}
	return nil
}

// Last 获取最后一个
func (m Members) Last() *User {
	if m.Count() > 0 {
		u := m[m.Count()-1]
		return u
	}
	return nil
}

// Append 追加联系人
func (m Members) Append(user *User) (results Members) {
	return append(m, user)
}

// SearchByUserName 根据用户名查找
func (m Members) SearchByUserName(limit int, username string) (results Members) {
	return m.Search(limit, func(user *User) bool { return user.UserName == username })
}

// SearchByNickName 根据昵称查找
func (m Members) SearchByNickName(limit int, nickName string) (results Members) {
	return m.Search(limit, func(user *User) bool { return user.NickName == nickName })
}

// SearchByRemarkName 根据备注查找
func (m Members) SearchByRemarkName(limit int, remarkName string) (results Members) {
	return m.Search(limit, func(user *User) bool { return user.RemarkName == remarkName })
}

// Search 根据自定义条件查找
func (m Members) Search(limit int, searchFuncList ...func(user *User) bool) (results Members) {
	return search(m, limit, func(group *User) bool {
		for _, searchFunc := range searchFuncList {
			if !searchFunc(group) {
				return false
			}
		}
		return true
	})
}

// GetByUserName 根据username查找用户
func (m Members) GetByUserName(username string) (*User, bool) {
	users := m.SearchByUserName(1, username)
	user := users.First()
	return user, user != nil
}

// GetByRemarkName 根据remarkName查找用户
func (m Members) GetByRemarkName(remarkName string) (*User, bool) {
	users := m.SearchByRemarkName(1, remarkName)
	user := users.First()
	return user, user != nil
}

// GetByNickName 根据nickname查找用户
func (m Members) GetByNickName(nickname string) (*User, bool) {
	users := m.SearchByNickName(1, nickname)
	user := users.First()
	return user, user != nil
}

func (m Members) Friends() Friends {
	friends := make(Friends, 0)
	for _, mb := range m {
		friend, ok := mb.AsFriend()
		if ok {
			friends = append(friends, friend)
		}
	}
	return friends
}

func (m Members) Groups() Groups {
	groups := make(Groups, 0)
	for _, mb := range m {
		group, ok := mb.AsGroup()
		if ok {
			groups = append(groups, group)
		}
	}
	return groups
}

func (m Members) MPs() Mps {
	mps := make(Mps, 0)
	for _, mb := range m {
		mp, ok := mb.AsMP()
		if ok {
			mps = append(mps, mp)
		}
	}
	return mps
}

type membersUpdater struct {
	self       *Self
	members    Members
	max        int
	index      int
	updateTime int
	current    Members
}

func (m *membersUpdater) init() {
	if m.members.Count() > 0 {
		m.self = m.members.First().Self()
	}
	if m.members.Count() <= m.max {
		m.updateTime = 1
	} else {
		m.updateTime = m.members.Count() / m.max
		if m.members.Count()%m.max != 0 {
			m.updateTime++
		}
	}
}

func (m *membersUpdater) Next() bool {
	if m.index >= m.updateTime {
		return false
	}
	m.index++
	return true
}

func (m *membersUpdater) Update() error {
	start := m.max * (m.index - 1)

	end := m.max * m.index

	if m.index == m.updateTime {
		end = m.members.Count()
	}

	// 获取需要更新的联系人
	m.current = m.members[start:end]
	ctx := m.self.Bot().Context()
	req := m.self.Bot().Storage.Request
	members, err := m.self.Bot().Caller.WebWxBatchGetContact(ctx, m.current, req)
	if err != nil {
		return err
	}

	// 更新联系人
	for i, member := range members {
		member.self = m.self
		member.formatEmoji()
		m.members[start+i] = member
	}
	return nil
}

func newMembersUpdater(members Members) *membersUpdater {
	return &membersUpdater{
		members: members,
		max:     50,
	}
}

// Detail 获取当前 Members 的详情
func (m Members) Detail() error {
	if m.Count() == 0 {
		return nil
	}
	updater := newMembersUpdater(m)
	updater.init()
	for updater.Next() {
		if err := updater.Update(); err != nil {
			return err
		}
	}
	return nil
}

func (m Members) init(self *Self) {
	for _, member := range m {
		member.self = self
		member.formatEmoji()
	}
}

func newFriend(username string, self *Self) *Friend {
	return &Friend{User: newUser(self, username)}
}

// NewFriendHelper 创建一个文件传输助手
// 文件传输助手的微信身份标识符永远是filehelper
func NewFriendHelper(self *Self) *Friend {
	return newFriend(FileHelper, self)
}

// SendTextToMp 发送文本消息给公众号
func (s *Self) SendTextToMp(mp *Mp, text string) (*SentMessage, error) {
	return s.sendTextToUser(mp.User.UserName, text)
}

// SendImageToMp 发送图片消息给公众号
func (s *Self) SendImageToMp(mp *Mp, file io.Reader) (*SentMessage, error) {
	return s.sendImageToUser(mp.User.UserName, file)
}

// SendFileToMp 发送文件给公众号
func (s *Self) SendFileToMp(mp *Mp, file io.Reader) (*SentMessage, error) {
	return s.sendFileToUser(mp.User.UserName, file)
}

// SendVideoToMp 发送视频消息给公众号
func (s *Self) SendVideoToMp(mp *Mp, file io.Reader) (*SentMessage, error) {
	return s.sendVideoToUser(mp.User.UserName, file)
}

func (s *Self) sendMessageWrapper(message *SentMessage, err error) (*SentMessage, error) {
	if err != nil {
		return nil, err
	}
	message.self = s
	return message, nil
}

// Bot 获取当前用户的机器人
func (s *Self) Bot() *Bot {
	return s.bot
}
