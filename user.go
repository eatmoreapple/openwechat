package openwechat

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// User 抽象的用户结构: 好友 群组 公众号
type User struct {
	Uin               int
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
	AttrStatus        int
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

	Self *Self
}

// implement fmt.Stringer
func (u *User) String() string {
	return fmt.Sprintf("<User:%s>", u.NickName)
}

// GetAvatarResponse 获取用户头像
func (u *User) GetAvatarResponse() (*http.Response, error) {
	return u.Self.Bot.Caller.Client.WebWxGetHeadImg(u.HeadImgUrl)
}

// SaveAvatar 下载用户头像
func (u *User) SaveAvatar(filename string) error {
	resp, err := u.GetAvatarResponse()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buffer := bytes.Buffer{}
	if _, err := buffer.ReadFrom(resp.Body); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(buffer.Bytes())
	return err
}

// Detail 获取用户的详情
func (u *User) Detail() (*User, error) {
	if u.UserName == u.Self.UserName {
		return u.Self.User, nil
	}
	members := Members{u}
	request := u.Self.Bot.Storage.Request
	newMembers, err := u.Self.Bot.Caller.WebWxBatchGetContact(members, request)
	if err != nil {
		return nil, err
	}
	newMembers.init(u.Self)
	user := newMembers.First()
	return user, nil
}

// IsFriend 判断是否为好友
func (u *User) IsFriend() bool {
	return !u.IsGroup() && strings.HasPrefix(u.UserName, "@") && u.VerifyFlag == 0
}

// IsGroup 判断是否为群组
func (u *User) IsGroup() bool {
	return strings.HasPrefix(u.UserName, "@@") && u.VerifyFlag == 0
}

// IsMP  判断是否为公众号
func (u *User) IsMP() bool {
	return u.VerifyFlag == 8 || u.VerifyFlag == 24 || u.VerifyFlag == 136
}

// Pin 将联系人置顶
func (u *User) Pin() error {
	req := u.Self.Bot.Storage.Request
	return u.Self.Bot.Caller.WebWxRelationPin(req, u, 1)
}

// UnPin 将联系人取消置顶
func (u *User) UnPin() error {
	req := u.Self.Bot.Storage.Request
	return u.Self.Bot.Caller.WebWxRelationPin(req, u, 0)
}

// IsPin 判断当前联系人(好友、群组、公众号)是否为置顶状态
func (u *User) IsPin() bool {
	return u.ContactFlag == 2051
}

// Self 自己,当前登录用户对象
type Self struct {
	*User
	Bot        *Bot
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
	return s.members, nil
}

// 更新联系人处理
func (s *Self) updateMembers() error {
	info := s.Bot.Storage.LoginInfo
	members, err := s.Bot.Caller.WebWxGetContact(info)
	if err != nil {
		return err
	}
	members.init(s)
	s.members = members
	return nil
}

// FileHelper 获取文件传输助手对象，封装成Friend返回
//      fh, err := self.FileHelper() // or fh := openwechat.NewFriendHelper(self)
func (s *Self) FileHelper() (*Friend, error) {
	// 如果缓存里有，直接返回，否则去联系人里面找
	if s.fileHelper != nil {
		return s.fileHelper, nil
	}
	members, err := s.Members()
	if err != nil {
		return nil, err
	}
	users := members.SearchByUserName(1, "filehelper")
	if users == nil {
		s.fileHelper = NewFriendHelper(s)
	} else {
		s.fileHelper = &Friend{users.First()}
	}
	return s.fileHelper, nil
}

// Friends 获取所有的好友
func (s *Self) Friends(update ...bool) (Friends, error) {
	if s.friends == nil || (len(update) > 0 && update[0]) {
		if _, err := s.Members(true); err != nil {
			return nil, err
		}
		s.friends = s.members.Friends()
	}
	return s.friends, nil
}

// Groups 获取所有的群组
func (s *Self) Groups(update ...bool) (Groups, error) {
	if s.groups == nil || (len(update) > 0 && update[0]) {
		if _, err := s.Members(true); err != nil {
			return nil, err
		}
		s.groups = s.members.Groups()
	}
	return s.groups, nil
}

// Mps 获取所有的公众号
func (s *Self) Mps(update ...bool) (Mps, error) {
	if s.mps == nil || (len(update) > 0 && update[0]) {
		if _, err := s.Members(true); err != nil {
			return nil, err
		}
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
	return members.detail(s)
}

// 抽象发送消息接口
func (s *Self) sendMessageToUser(user *User, msg *SendMessage) (*SentMessage, error) {
	msg.FromUserName = s.UserName
	msg.ToUserName = user.UserName
	info := s.Bot.Storage.LoginInfo
	request := s.Bot.Storage.Request
	successSendMessage, err := s.Bot.Caller.WebWxSendMsg(msg, info, request)
	if err != nil {
		return nil, err
	}
	successSendMessage.Self = s
	return successSendMessage, nil
}

// SendMessageToFriend 发送消息给好友
func (s *Self) SendMessageToFriend(friend *Friend, msg *SendMessage) (*SentMessage, error) {
	return s.sendMessageToUser(friend.User, msg)
}

// SendTextToFriend 发送文本消息给好友
func (s *Self) SendTextToFriend(friend *Friend, text string) (*SentMessage, error) {
	msg := NewTextSendMessage(text, s.UserName, friend.UserName)
	return s.SendMessageToFriend(friend, msg)
}

// SendImageToFriend 发送图片消息给好友
func (s *Self) SendImageToFriend(friend *Friend, file *os.File) (*SentMessage, error) {
	req := s.Bot.Storage.Request
	info := s.Bot.Storage.LoginInfo
	return s.Bot.Caller.WebWxSendImageMsg(file, req, info, s.UserName, friend.UserName)
}

// SendFileToFriend 发送文件给好友
func (s *Self) SendFileToFriend(friend *Friend, file *os.File) (*SentMessage, error) {
	req := s.Bot.Storage.Request
	info := s.Bot.Storage.LoginInfo
	return s.Bot.Caller.WebWxSendFile(file, req, info, s.UserName, friend.UserName)
}

// SetRemarkNameToFriend 设置好友备注
//      self.SetRemarkNameToFriend(friend, "remark") // or friend.SetRemarkName("remark")
func (s *Self) SetRemarkNameToFriend(friend *Friend, remarkName string) error {
	req := s.Bot.Storage.Request
	return s.Bot.Caller.WebWxOplog(req, remarkName, friend.UserName)
}

// AddFriendsIntoGroup 拉多名好友进群
// 最好自己是群主,成功率高一点,因为有的群允许非群组拉人,而有的群不允许
func (s *Self) AddFriendsIntoGroup(group *Group, friends ...*Friend) error {
	if len(friends) == 0 {
		return nil
	}
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
	req := s.Bot.Storage.Request
	info := s.Bot.Storage.LoginInfo
	return s.Bot.Caller.AddFriendIntoChatRoom(req, info, group, friends...)
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
	req := s.Bot.Storage.Request
	info := s.Bot.Storage.LoginInfo
	return s.Bot.Caller.RemoveFriendFromChatRoom(req, info, group, members...)
}

// AddFriendIntoManyGroups 拉好友进多个群聊
// AddFriendIntoGroups, 名字和上面的有点像
func (s *Self) AddFriendIntoManyGroups(friend *Friend, groups ...*Group) error {
	for _, group := range groups {
		if err := s.AddFriendsIntoGroup(group, friend); err != nil {
			return err
		}
	}
	return nil
}

// SendMessageToGroup 发送消息给群组
func (s *Self) SendMessageToGroup(group *Group, msg *SendMessage) (*SentMessage, error) {
	return s.sendMessageToUser(group.User, msg)
}

// SendTextToGroup 发送文本消息给群组
func (s *Self) SendTextToGroup(group *Group, text string) (*SentMessage, error) {
	msg := NewTextSendMessage(text, s.UserName, group.UserName)
	return s.SendMessageToGroup(group, msg)
}

// SendImageToGroup 发送图片消息给群组
func (s *Self) SendImageToGroup(group *Group, file *os.File) (*SentMessage, error) {
	req := s.Bot.Storage.Request
	info := s.Bot.Storage.LoginInfo
	return s.Bot.Caller.WebWxSendImageMsg(file, req, info, s.UserName, group.UserName)
}

// SendFileToGroup 发送文件给群组
func (s *Self) SendFileToGroup(group *Group, file *os.File) (*SentMessage, error) {
	req := s.Bot.Storage.Request
	info := s.Bot.Storage.LoginInfo
	return s.Bot.Caller.WebWxSendFile(file, req, info, s.UserName, group.UserName)
}

// RevokeMessage 撤回消息
//      sentMessage, err := friend.SendText("message")
//      if err == nil {
//          self.RevokeMessage(sentMessage) // or sentMessage.Revoke()
//      }
func (s *Self) RevokeMessage(msg *SentMessage) error {
	return s.Bot.Caller.WebWxRevokeMsg(msg, s.Bot.Storage.Request)
}

// 转发消息接口
func (s *Self) forwardMessage(msg *SentMessage, users ...*User) error {
	info := s.Bot.Storage.LoginInfo
	req := s.Bot.Storage.Request
	switch msg.Type {
	case MsgTypeText:
		for _, user := range users {
			msg.FromUserName = s.UserName
			msg.ToUserName = user.UserName
			if _, err := s.Self.Bot.Caller.WebWxSendMsg(msg.SendMessage, info, req); err != nil {
				return err
			}
		}
	case MsgTypeImage:
		for _, user := range users {
			msg.FromUserName = s.UserName
			msg.ToUserName = user.UserName
			if _, err := s.Self.Bot.Caller.Client.WebWxSendMsgImg(msg.SendMessage, req, info); err != nil {
				return err
			}
		}
	case AppMessage:
		for _, user := range users {
			msg.FromUserName = s.UserName
			msg.ToUserName = user.UserName
			if _, err := s.Self.Bot.Caller.Client.WebWxSendAppMsg(msg.SendMessage, req); err != nil {
				return err
			}
		}
	}
	return errors.New("unsupport message")
}

// ForwardMessageToFriends 转发给好友
func (s *Self) ForwardMessageToFriends(msg *SentMessage, friends ...*Friend) error {
	var users = make([]*User, len(friends))
	for index, friend := range friends {
		users[index] = friend.User
	}
	return s.forwardMessage(msg, users...)
}

// ForwardMessageToGroups 转发给群组
func (s *Self) ForwardMessageToGroups(msg *SentMessage, groups ...*Group) error {
	var users = make([]*User, len(groups))
	for index, group := range groups {
		users[index] = group.User
	}
	return s.forwardMessage(msg, users...)
}

// Members 抽象的用户组
type Members []*User

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
func (m Members) Search(limit int, condFuncList ...func(user *User) bool) (results Members) {
	if condFuncList == nil {
		return m
	}
	if limit <= 0 {
		limit = m.Count()
	}
	for _, member := range m {
		if count := len(results); count == limit {
			break
		}
		var passCount int
		for _, condFunc := range condFuncList {
			if condFunc(member) {
				passCount++
			} else {
				break
			}
		}
		if passCount == len(condFuncList) {
			results = append(results, member)
		}
	}
	return
}

func (m Members) Friends() Friends {
	friends := make(Friends, 0)
	for _, mb := range m {
		if mb.IsFriend() {
			friend := &Friend{mb}
			friends = append(friends, friend)
		}
	}
	return friends
}

func (m Members) Groups() Groups {
	groups := make(Groups, 0)
	for _, mb := range m {
		if mb.IsGroup() {
			group := &Group{mb}
			groups = append(groups, group)
		}
	}
	return groups
}

func (m Members) MPs() Mps {
	mps := make(Mps, 0)
	for _, mb := range m {
		if mb.IsMP() {
			mp := &Mp{mb}
			mps = append(mps, mp)
		}
	}
	return mps
}

// 获取当前Members的详情
func (m Members) detail(self *Self) error {
	// 获取他们的数量
	members := m

	count := members.Count()
	// 一次更新50个,分情况讨论

	// 获取总的需要更新的次数
	var times int
	if count < 50 {
		times = 1
	} else {
		times = count / 50
	}
	var newMembers Members
	request := self.Bot.Storage.Request
	var pMembers Members
	// 分情况依次更新
	for i := 1; i <= times; i++ {
		if times == 1 {
			pMembers = members
		} else {
			pMembers = members[(i-1)*50 : i*50]
		}
		nMembers, err := self.Bot.Caller.WebWxBatchGetContact(pMembers, request)
		if err != nil {
			return err
		}
		newMembers = append(newMembers, nMembers...)
	}
	// 最后判断是否全部更新完毕
	total := times * 50
	if total < count {
		// 将全部剩余的更新完毕
		left := count - total
		pMembers = members[total : total+left]
		nMembers, err := self.Bot.Caller.WebWxBatchGetContact(pMembers, request)
		if err != nil {
			return err
		}
		newMembers = append(newMembers, nMembers...)
	}
	if len(newMembers) > 0 {
		newMembers.init(self)
		self.members = newMembers
	}
	return nil
}

func (m Members) init(self *Self) {
	for _, member := range m {
		member.Self = self
		member.NickName = FormatEmoji(member.NickName)
		member.RemarkName = FormatEmoji(member.RemarkName)
		member.DisplayName = FormatEmoji(member.DisplayName)
	}
}

// NewFriendHelper 这里为了兼容Desktop版本找不到文件传输助手的问题
// 文件传输助手的微信身份标识符永远是filehelper
// 这种形式的对象可能缺少一些其他属性
// 但是不影响发送信息的功能
func NewFriendHelper(self *Self) *Friend {
	return &Friend{&User{UserName: "filehelper", Self: self}}
}
