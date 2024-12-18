package openwechat

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

type Friend struct{ *User }

// implement fmt.Stringer
func (f *Friend) String() string {
	display := f.NickName
	if f.RemarkName != "" {
		display = f.RemarkName
	}
	return fmt.Sprintf("<Friend:%s>", display)
}

// SetRemarkName 重命名当前好友
// Deprecated
func (f *Friend) SetRemarkName(name string) error {
	return f.Self().SetRemarkNameToFriend(f, name)
}

// SendText  发送文本消息
func (f *Friend) SendText(content string) (*SentMessage, error) {
	return f.Self().SendTextToFriend(f, content)
}

// SendImage 发送图片消息
func (f *Friend) SendImage(file io.Reader) (*SentMessage, error) {
	return f.Self().SendImageToFriend(f, file)
}

// SendVideo 发送视频消息
func (f *Friend) SendVideo(file io.Reader) (*SentMessage, error) {
	return f.Self().SendVideoToFriend(f, file)
}

// SendFile 发送文件消息
func (f *Friend) SendFile(file io.Reader) (*SentMessage, error) {
	return f.Self().SendFileToFriend(f, file)
}

// AddIntoGroup 拉该好友入群
func (f *Friend) AddIntoGroup(groups ...*Group) error {
	return f.Self().AddFriendIntoManyGroups(f, groups...)
}

type Friends []*Friend

// Count 获取好友的数量
func (f Friends) Count() int {
	return len(f)
}

// First 获取第一个好友
func (f Friends) First() *Friend {
	if f.Count() > 0 {
		return f.Sort()[0]
	}
	return nil
}

// Last 获取最后一个好友
func (f Friends) Last() *Friend {
	if f.Count() > 0 {
		return f.Sort()[f.Count()-1]
	}
	return nil
}

// SearchByUserName 根据用户名查找好友
func (f Friends) SearchByUserName(limit int, username string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.UserName == username })
}

// SearchByNickName 根据昵称查找好友
func (f Friends) SearchByNickName(limit int, nickName string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.NickName == nickName })
}

// SearchByRemarkName 根据备注查找好友
func (f Friends) SearchByRemarkName(limit int, remarkName string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.RemarkName == remarkName })
}

// SearchByID 根据ID查找好友
func (f Friends) SearchByID(id string) Friends {
	return f.Search(1, func(friend *Friend) bool { return friend.User.ID() == id })
}

// Search 根据自定义条件查找好友
func (f Friends) Search(limit int, searchFuncList ...func(friend *Friend) bool) (results Friends) {
	return f.AsMembers().Search(limit, func(user *User) bool {
		var friend = &Friend{user}
		for _, searchFunc := range searchFuncList {
			if !searchFunc(friend) {
				return false
			}
		}
		return true
	}).Friends()
}

// AsMembers 将群组转换为用户列表
func (f Friends) AsMembers() Members {
	var members = make(Members, 0, f.Count())
	for _, friend := range f {
		members = append(members, friend.User)
	}
	return members
}

// Sort 对好友进行排序
func (f Friends) Sort() Friends {
	return f.AsMembers().Sort().Friends()
}

// Uniq 对好友进行去重
func (f Friends) Uniq() Friends {
	return f.AsMembers().Uniq().Friends()
}

// SendText 向slice的好友依次发送文本消息
func (f Friends) SendText(text string, delays ...time.Duration) error {
	if f.Count() == 0 {
		return nil
	}
	var delay time.Duration
	if len(delays) > 0 {
		delay = delays[0]
	}
	self := f.First().Self()
	return self.SendTextToFriends(text, delay, f...)
}

// BroadcastTextToFriendsByRandomTime 向所有好友随机时间间隔发送消息。
func (f Friends) BroadcastTextToFriendsByRandomTime(msg string) error {
	for _, friend := range f {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second) //随机休眠0-10秒
		if _, err := friend.SendText(msg); err != nil {
			return err
		}
	}
	return nil
}

// SendImage 向slice的好友依次发送图片消息
func (f Friends) SendImage(file io.Reader, delays ...time.Duration) error {
	if f.Count() == 0 {
		return nil
	}
	var delay time.Duration
	if len(delays) > 0 {
		delay = delays[0]
	}
	self := f.First().Self()
	return self.SendImageToFriends(file, delay, f...)
}

// SendFile 群发文件
func (f Friends) SendFile(file io.Reader, delay ...time.Duration) error {
	if f.Count() == 0 {
		return nil
	}
	var d time.Duration
	if len(delay) > 0 {
		d = delay[0]
	}
	self := f.First().Self()
	return self.SendFileToFriends(file, d, f...)
}

type Group struct{ *User }

// implement fmt.Stringer
func (g *Group) String() string {
	return fmt.Sprintf("<Group:%s>", g.NickName)
}

// SendText 发送文本消息给当前的群组
func (g *Group) SendText(content string) (*SentMessage, error) {
	return g.Self().SendTextToGroup(g, content)
}

// SendImage 发送图片消息给当前的群组
func (g *Group) SendImage(file io.Reader) (*SentMessage, error) {
	return g.Self().SendImageToGroup(g, file)
}

// SendEmoticon 发送表情消息给当前的群组
func (g *Group) SendEmoticon(md5 string, file io.Reader) (*SentMessage, error) {
	return g.Self().SendEmoticonToGroup(g, md5, file)
}

// SendVideo 发送视频消息给当前的群组
func (g *Group) SendVideo(file io.Reader) (*SentMessage, error) {
	return g.Self().SendVideoToGroup(g, file)
}

// SendFile 发送文件给当前的群组
func (g *Group) SendFile(file io.Reader) (*SentMessage, error) {
	return g.Self().SendFileToGroup(g, file)
}

// Members 获取所有的群成员
func (g *Group) Members() (Members, error) {
	if err := g.Detail(); err != nil {
		return nil, err
	}
	g.MemberList.init(g.Self())
	return g.MemberList, nil
}

// AddFriendsIn 拉好友入群
func (g *Group) AddFriendsIn(friends ...*Friend) error {
	friends = Friends(friends).Uniq()
	return g.self.AddFriendsIntoGroup(g, friends...)
}

// RemoveMembers 从群聊中移除用户
// Deprecated
// 无论是网页版，还是程序上都不起作用
func (g *Group) RemoveMembers(members Members) error {
	return g.Self().RemoveMemberFromGroup(g, members)
}

// Rename 群组重命名
// Deprecated
func (g *Group) Rename(name string) error {
	return g.Self().RenameGroup(g, name)
}

// SearchMemberByUsername 根据用户名查找群成员
func (g *Group) SearchMemberByUsername(username string) (*User, error) {
	if g.MemberList.Count() == 0 {
		if _, err := g.Members(); err != nil {
			return nil, err
		}
	}
	members := g.MemberList.SearchByUserName(1, username)
	// 如果此时本地查不到, 那么该成员可能是新加入的
	if members.Count() == 0 {
		if _, err := g.Members(); err != nil {
			return nil, err
		}
	}
	// 再次尝试获取
	members = g.MemberList.SearchByUserName(1, username)
	if members.Count() == 0 {
		return nil, ErrNoSuchUserFound
	}
	return members.First(), nil
}

type Groups []*Group

// Count 获取群组数量
func (g Groups) Count() int {
	return len(g)
}

// First 获取第一个群组
func (g Groups) First() *Group {
	if g.Count() > 0 {
		return g.Sort()[0]
	}
	return nil
}

// Last 获取最后一个群组
func (g Groups) Last() *Group {
	if g.Count() > 0 {
		return g.Sort()[g.Count()-1]
	}
	return nil
}

// SendText 向群组依次发送文本消息, 支持发送延迟
func (g Groups) SendText(text string, delay ...time.Duration) error {
	if g.Count() == 0 {
		return nil
	}
	var d time.Duration
	if len(delay) > 0 {
		d = delay[0]
	}
	self := g.First().Self()
	return self.SendTextToGroups(text, d, g...)
}

// SendImage 向群组依次发送图片消息, 支持发送延迟
func (g Groups) SendImage(file io.Reader, delay ...time.Duration) error {
	if g.Count() == 0 {
		return nil
	}
	var d time.Duration
	if len(delay) > 0 {
		d = delay[0]
	}
	self := g.First().Self()
	return self.SendImageToGroups(file, d, g...)
}

// SendFile 向群组依次发送文件消息, 支持发送延迟
func (g Groups) SendFile(file io.Reader, delay ...time.Duration) error {
	if g.Count() == 0 {
		return nil
	}
	var d time.Duration
	if len(delay) > 0 {
		d = delay[0]
	}
	self := g.First().Self()
	return self.SendFileToGroups(file, d, g...)
}

// SearchByUserName 根据用户名查找群组
func (g Groups) SearchByUserName(limit int, username string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.UserName == username })
}

// SearchByNickName 根据昵称查找群组
func (g Groups) SearchByNickName(limit int, nickName string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.NickName == nickName })
}

// SearchByID 根据ID查找群组
func (g Groups) SearchByID(id string) Groups {
	return g.Search(1, func(group *Group) bool { return group.ID() == id })
}

// Search 根据自定义条件查找群组
func (g Groups) Search(limit int, searchFuncList ...func(group *Group) bool) (results Groups) {
	return g.AsMembers().Search(limit, func(user *User) bool {
		var group = &Group{user}
		for _, searchFunc := range searchFuncList {
			if !searchFunc(group) {
				return false
			}
		}
		return true
	}).Groups()
}

// AsMembers 将群组列表转换为用户列表
func (g Groups) AsMembers() Members {
	var members = make(Members, 0, g.Count())
	for _, group := range g {
		members = append(members, group.User)
	}
	return members
}

// Sort 对群组进行排序
func (g Groups) Sort() Groups {
	return g.AsMembers().Sort().Groups()
}

// Uniq 对群组进行去重
func (g Groups) Uniq() Groups {
	return g.AsMembers().Uniq().Groups()
}

// Mp 公众号对象
type Mp struct{ *User }

func (m *Mp) String() string {
	return fmt.Sprintf("<Mp:%s>", m.NickName)
}

// Mps 公众号组对象
type Mps []*Mp

// Count 数量统计
func (m Mps) Count() int {
	return len(m)
}

// First 获取第一个
func (m Mps) First() *Mp {
	if m.Count() > 0 {
		return m.Sort()[0]
	}
	return nil
}

// Last 获取最后一个
func (m Mps) Last() *Mp {
	if m.Count() > 0 {
		return m.Sort()[m.Count()-1]
	}
	return nil
}

// Search 根据自定义条件查找
func (m Mps) Search(limit int, searchFuncList ...func(*Mp) bool) (results Mps) {
	return m.AsMembers().Search(limit, func(user *User) bool {
		var mp = &Mp{user}
		for _, searchFunc := range searchFuncList {
			if !searchFunc(mp) {
				return false
			}
		}
		return true
	}).MPs()
}

// AsMembers 将公众号列表转换为用户列表
func (m Mps) AsMembers() Members {
	var members = make(Members, 0, m.Count())
	for _, mp := range m {
		members = append(members, mp.User)
	}
	return members
}

// Sort 对公众号进行排序
func (m Mps) Sort() Mps {
	return m.AsMembers().Sort().MPs()
}

// Uniq 对公众号进行去重
func (m Mps) Uniq() Mps {
	return m.AsMembers().Uniq().MPs()
}

// SearchByUserName 根据用户名查找
func (m Mps) SearchByUserName(limit int, userName string) (results Mps) {
	return m.Search(limit, func(mp *Mp) bool { return mp.UserName == userName })
}

// SearchByNickName 根据昵称查找
func (m Mps) SearchByNickName(limit int, nickName string) (results Mps) {
	return m.Search(limit, func(mp *Mp) bool { return mp.NickName == nickName })
}

// SendText 发送文本消息给公众号
func (m *Mp) SendText(content string) (*SentMessage, error) {
	return m.Self().SendTextToMp(m, content)
}

// SendImage 发送图片消息给公众号
func (m *Mp) SendImage(file io.Reader) (*SentMessage, error) {
	return m.Self().SendImageToMp(m, file)
}

// SendFile 发送文件消息给公众号
func (m *Mp) SendFile(file io.Reader) (*SentMessage, error) {
	return m.Self().SendFileToMp(m, file)
}

// GetByUsername 根据username查询一个Friend
func (f Friends) GetByUsername(username string) *Friend {
	return f.SearchByUserName(1, username).First()
}

// GetByRemarkName 根据remarkName查询一个Friend
func (f Friends) GetByRemarkName(remarkName string) *Friend {
	return f.SearchByRemarkName(1, remarkName).First()
}

// GetByNickName 根据nickname查询一个Friend
func (f Friends) GetByNickName(nickname string) *Friend {
	return f.SearchByNickName(1, nickname).First()
}

// GetByUsername 根据username查询一个Group
func (g Groups) GetByUsername(username string) *Group {
	return g.SearchByUserName(1, username).First()
}

// GetByNickName 根据nickname查询一个Group
func (g Groups) GetByNickName(nickname string) *Group {
	return g.SearchByNickName(1, nickname).First()
}

// GetByNickName 根据nickname查询一个Mp
func (m Mps) GetByNickName(nickname string) *Mp {
	return m.SearchByNickName(1, nickname).First()
}

// GetByUserName 根据username查询一个Mp
func (m Mps) GetByUserName(username string) *Mp {
	return m.SearchByUserName(1, username).First()
}

// search 根据自定义条件查找
func search(searchList Members, limit int, searchFunc func(*User) bool) (results Members) {
	if limit <= 0 {
		limit = searchList.Count()
	}
	for _, member := range searchList {
		if results.Count() == limit {
			break
		}
		if searchFunc(member) {
			results = append(results, member)
		}
	}
	return
}
