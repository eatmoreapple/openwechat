package openwechat

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

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
	UserName          string
	NickName          string
	HeadImgUrl        string
	RemarkName        string
	PYInitial         string
	PYQuanPin         string
	RemarkPYInitial   string
	RemarkPYQuanPin   string
	Signature         string
	MemberCount       int
	MemberList        []*User
	OwnerUin          int
	Statues           int
	AttrStatus        int
	Province          string
	City              string
	Alias             string
	UniFriend         int
	DisplayName       string
	ChatRoomId        int
	KeyWord           string
	EncryChatRoomId   string
	IsOwner           int

	Self *Self
}

// implement fmt.Stringer
func (u *User) String() string {
	return fmt.Sprintf("<User:%s>", u.NickName)
}

// 获取用户头像
func (u *User) GetAvatarResponse() (*http.Response, error) {
	return u.Self.Bot.Caller.Client.WebWxGetHeadImg(u.HeadImgUrl)
}

// 下载用户头像
func (u *User) SaveAvatar(filename string) error {
	resp, err := u.GetAvatarResponse()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, os.ModePerm)
}

func (u *User) sendMsg(msg *SendMessage) error {
	msg.FromUserName = u.Self.UserName
	msg.ToUserName = u.UserName
	info := u.Self.Bot.storage.GetLoginInfo()
	request := u.Self.Bot.storage.GetBaseRequest()
	return u.Self.Bot.Caller.WebWxSendMsg(msg, info, request)
}

func (u *User) sendText(content string) error {
	msg := NewTextSendMessage(content, u.Self.UserName, u.UserName)
	return u.sendMsg(msg)
}

func (u *User) sendImage(file *os.File) error {
	request := u.Self.Bot.storage.GetBaseRequest()
	info := u.Self.Bot.storage.GetLoginInfo()
	return u.Self.Bot.Caller.WebWxSendImageMsg(file, request, info, u.Self.UserName, u.UserName)
}

func (u *User) remakeName(remarkName string) error {
	request := u.Self.Bot.storage.GetBaseRequest()
	return u.Self.Bot.Caller.WebWxOplog(request, remarkName, u.UserName)
}

// 获取用户的详情
func (u *User) Detail() (*User, error) {
	members := Members{u}
	request := u.Self.Bot.storage.GetBaseRequest()
	newMembers, err := u.Self.Bot.Caller.WebWxBatchGetContact(members, request)
	if err != nil {
		return nil, err
	}
	user := newMembers[0]
	user.Self = u.Self
	return user, nil
}

type Self struct {
	*User
	Bot        *Bot
	fileHelper *Friend
	members    Members
	friends    Friends
	groups     Groups
}

// 获取所有的好友、群组、公众号信息
func (s *Self) Members(update ...bool) (Members, error) {
	// 首先判断缓存里有没有,如果没有则去更新缓存
	if s.members == nil {
		if err := s.updateMembers(); err != nil {
			return nil, err
		}
		return s.members, nil
	}
	// 判断是否需要更新,如果传入的参数不为nil,则取最后一个
	var isUpdate bool
	if len(update) > 0 {
		isUpdate = update[len(update)-1]
	}
	// 如果需要更新，则直接更新缓存
	if isUpdate {
		if err := s.updateMembers(); err != nil {
			return nil, err
		}
	}
	return s.members, nil
}

// 更新联系人处理
func (s *Self) updateMembers() error {
	info := s.Bot.storage.GetLoginInfo()
	members, err := s.Bot.Caller.WebWxGetContact(info)
	if err != nil {
		return err
	}
	members.SetOwner(s)
	s.members = members
	return nil
}

// 获取文件传输助手对象，封装成Friend返回
func (s *Self) FileHelper() (*Friend, error) {
	// 如果缓存里有，直接返回，否则去联系人里面找
	if s.fileHelper != nil {
		return s.fileHelper, nil
	}
	members, err := s.Members()
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		if member.UserName == "filehelper" {
			fileHelper := &Friend{member}
			// 将找到的缓存起来,方便下次调用
			s.fileHelper = fileHelper
			return s.fileHelper, nil
		}
	}
	return nil, errors.New("filehelper does not exist")
}

// 获取所有的好友
func (s *Self) Friends(update ...bool) (Friends, error) {
	if s.friends == nil {
		if err := s.updateFriends(update...); err != nil {
			return nil, err
		}
	}
	return s.friends, nil
}

// 获取所有的群组
func (s *Self) Groups(update ...bool) (Groups, error) {
	if s.groups == nil {
		if err := s.updateGroups(update...); err != nil {
			return nil, err
		}
	}
	return s.groups, nil
}

// 更新好友处理
func (s *Self) updateFriends(update ...bool) error {
	var isUpdate bool
	if len(update) > 0 {
		isUpdate = update[len(update)-1]
	}
	if isUpdate || s.members == nil {
		if err := s.updateMembers(); err != nil {
			return err
		}
	}
	friends := make(Friends, 0)
	for _, member := range s.members {
		if isFriend(*member) {
			friend := &Friend{member}
			friends = append(friends, friend)
		}
	}
	s.friends = friends
	return nil
}

// 更新群组处理
func (s *Self) updateGroups(update ...bool) error {
	var isUpdate bool
	if len(update) > 0 {
		isUpdate = update[len(update)-1]
	}
	if isUpdate || s.members == nil {
		if err := s.updateMembers(); err != nil {
			return err
		}
	}
	groups := make(Groups, 0)
	for _, member := range s.members {
		if isGroup(*member) {
			group := &Group{member}
			groups = append(groups, group)
		}
	}
	s.groups = groups
	return nil
}

// 更新所有的联系人信息
func (s *Self) UpdateMembersDetail() error {
	// 先获取所有的联系人
	members, err := s.Members()
	if err != nil {
		return err
	}
	// 获取他们的数量
	count := members.Count()
	// 一次更新50个,分情况讨论

	// 获取总的需要更新的次数
	var times int
	if count < 50 {
		times = 1
	} else {
		times = count / 50
	}
	newMembers := make(Members, 0)
	request := s.Self.Bot.storage.GetBaseRequest()
	var pMembers Members
	// 分情况依次更新
	for i := 0; i < times; i++ {
		if times == 1 {
			pMembers = members
		} else {
			pMembers = members[i*50 : (i+1)*times]
		}
		nMembers, err := s.Self.Bot.Caller.WebWxBatchGetContact(pMembers, request)
		if err != nil {
			return err
		}
		newMembers = append(newMembers, nMembers...)
	}
	// 最后判断是否全部更新完毕
	total := times * 50
	if total < count {
		// 将全部剩余的更新完毕
		left := total - count
		pMembers = members[total : total+left]
		nMembers, err := s.Self.Bot.Caller.WebWxBatchGetContact(pMembers, request)
		if err != nil {
			return err
		}
		newMembers = append(newMembers, nMembers...)
	}
	newMembers.SetOwner(s)
	s.members = newMembers
	return nil
}

type Members []*User

func (m Members) Count() int {
	return len(m)
}

func (m Members) SetOwner(s *Self) {
	for _, member := range m {
		member.Self = s
	}
}
