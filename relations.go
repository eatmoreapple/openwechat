package openwechat

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Friend struct{ *User }

// implement fmt.Stringer
func (f Friend) String() string {
	return fmt.Sprintf("<Friend:%s>", f.NickName)
}

// 重命名当前好友
func (f *Friend) SetRemarkName(name string) error {
	return f.setRemarkName(name)
}

// 发送自定义消息
func (f *Friend) SendMsg(msg *SendMessage) error {
	return f.sendMsg(msg)
}

// 发送文本消息
func (f *Friend) SendText(content string) error {
	return f.sendText(content)
}

// 发送图片消息
func (f *Friend) SendImage(file *os.File) error {
	return f.sendImage(file)
}

// 拉该好友入群
func (f *Friend) AddIntoGroup(groups ...*Group) error {
	for _, group := range groups {
		if err := group.AddFriendsIn(f); err != nil {
			return err
		}
	}
	return nil
}

type Friends []*Friend

// 获取好友的数量
func (f Friends) Count() int {
	return len(f)
}

// 获取第一个好友
func (f Friends) First() *Friend {
	if f.Count() > 0 {
		return f[0]
	}
	return nil
}

// 获取最后一个好友
func (f Friends) Last() *Friend {
	if f.Count() > 0 {
		return f[f.Count()-1]
	}
	return nil
}

// 根据用户名查找好友
func (f Friends) SearchByUserName(limit int, username string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.UserName == username })
}

// 根据昵称查找好友
func (f Friends) SearchByNickName(limit int, nickName string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.NickName == nickName })
}

// 根据备注查找好友
func (f Friends) SearchByRemarkName(limit int, remarkName string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.RemarkName == remarkName })
}

// 根据自定义条件查找好友
func (f Friends) Search(limit int, condFuncList ...func(friend *Friend) bool) (results Friends) {
	if condFuncList == nil {
		return f
	}
	if limit <= 0 {
		limit = f.Count()
	}
	for _, member := range f {
		if results.Count() == limit {
			break
		}
		var passCount int
		for _, condFunc := range condFuncList {
			if condFunc(member) {
				passCount++
			}
		}
		if passCount == len(condFuncList) {
			results = append(results, member)
		}
	}
	return
}

// 向slice的好友依次发送消息
func (f Friends) SendMsg(msg *SendMessage, delay ...time.Duration) error {
	for _, friend := range f {
		if len(delay) != 0 {
			time.Sleep(delay[0])
		}
		if err := friend.SendMsg(msg); err != nil {
			return err
		}
	}
	return nil
}

// 向slice的好友依次发送文本消息
func (f Friends) SendText(text string, delay ...time.Duration) error {
	for _, friend := range f {
		if len(delay) != 0 {
			time.Sleep(delay[0])
		}
		if err := friend.SendText(text); err != nil {
			return err
		}
	}
	return nil
}

func (f Friends) SendImage(file *os.File, delay ...time.Duration) error {
	for _, friend := range f {
		if len(delay) != 0 {
			time.Sleep(delay[0])
		}
		if err := friend.SendImage(file); err != nil {
			return err
		}
	}
	return nil
}

type Group struct{ *User }

// implement fmt.Stringer
func (g Group) String() string {
	return fmt.Sprintf("<Group:%s>", g.NickName)
}

func (g *Group) SendMsg(msg *SendMessage) error {
	return g.sendMsg(msg)
}

func (g *Group) SendText(content string) error {
	return g.sendText(content)
}

func (g *Group) SendImage(file *os.File) error {
	return g.sendImage(file)
}

// 获取所有的群成员
func (g *Group) Members() (Members, error) {
	group, err := g.Detail()
	if err != nil {
		return nil, err
	}
	return group.MemberList, nil
}

// 拉好友入群
func (g *Group) AddFriendsIn(friends ...*Friend) error {
	if len(friends) == 0 {
		return nil
	}
	groupMembers, err := g.Members()
	if err != nil {
		return err
	}
	for _, friend := range friends {
		for _, member := range groupMembers {
			if member.UserName == friend.UserName {
				return fmt.Errorf("user %s has alreay in this group", friend.String())
			}
		}
	}
	req := g.Self.Bot.storage.Request
	info := g.Self.Bot.storage.LoginInfo
	return g.Self.Bot.Caller.AddFriendIntoChatRoom(req, info, g, friends...)
}

// 从群聊中移除用户
// Deprecated
// 无论是网页版，还是程序上都不起作用
func (g *Group) RemoveMembers(members Members) error {
	if len(members) == 0 {
		return nil
	}
	if g.IsOwner == 0 {
		return errors.New("group owner required")
	}
	groupMembers, err := g.Members()
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
	req := g.Self.Bot.storage.Request
	info := g.Self.Bot.storage.LoginInfo
	return g.Self.Bot.Caller.RemoveFriendFromChatRoom(req, info, g, members...)
}

type Groups []*Group

func (g Groups) Count() int {
	return len(g)
}

func (g Groups) First() *Group {
	if g.Count() > 0 {
		return g[0]
	}
	return nil
}

func (g Groups) Last() *Group {
	if g.Count() > 0 {
		return g[g.Count()-1]
	}
	return nil
}

func (g Groups) SendMsg(msg *SendMessage, delay ...time.Duration) error {
	for _, group := range g {
		if len(delay) != 0 {
			time.Sleep(delay[0])
		}
		if err := group.SendMsg(msg); err != nil {
			return err
		}
	}
	return nil
}

func (g Groups) SendText(text string, delay ...time.Duration) error {
	for _, group := range g {
		if len(delay) != 0 {
			time.Sleep(delay[0])
		}
		if err := group.SendText(text); err != nil {
			return err
		}
	}
	return nil
}

func (g Groups) SendImage(file *os.File, delay ...time.Duration) error {
	for _, group := range g {
		if len(delay) != 0 {
			time.Sleep(delay[0])
		}
		if err := group.SendImage(file); err != nil {
			return err
		}
	}
	return nil
}

func (g Groups) SearchByUserName(limit int, username string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.UserName == username })
}

func (g Groups) SearchByNickName(limit int, nickName string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.NickName == nickName })
}

func (g Groups) SearchByRemarkName(limit int, remarkName string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.RemarkName == remarkName })
}

func (g Groups) Search(limit int, condFuncList ...func(group *Group) bool) (results Groups) {
	if condFuncList == nil {
		return g
	}
	if limit <= 0 {
		limit = g.Count()
	}
	for _, member := range g {
		if results.Count() == limit {
			break
		}
		var passCount int
		for _, condFunc := range condFuncList {
			if condFunc(member) {
				passCount++
			}
		}
		if passCount == len(condFuncList) {
			results = append(results, member)
		}
	}
	return
}

func isFriend(user User) bool {
	return !isGroup(user) && strings.HasPrefix(user.UserName, "@") && user.VerifyFlag == 0
}

func isGroup(user User) bool {
	return strings.HasPrefix(user.UserName, "@@") && user.VerifyFlag == 0
}

func isMP(user User) bool {
	return user.VerifyFlag == 8 || user.VerifyFlag == 24 || user.VerifyFlag == 136
}

type Mp struct{ *User }

func (m Mp) String() string {
	return fmt.Sprintf("<Mp:%s>", m.NickName)
}

type Mps []*Mp

func (m Mps) Count() int {
	return len(m)
}

func (m Mps) First() *Mp {
	if m.Count() > 0 {
		return m[0]
	}
	return nil
}

func (m Mps) Last() *Mp {
	if m.Count() > 0 {
		return m[m.Count()-1]
	}
	return nil
}

func (m Mps) Search(limit int, condFuncList ...func(group *Mp) bool) (results Mps) {
	if condFuncList == nil {
		return m
	}
	if limit <= 0 {
		limit = m.Count()
	}
	for _, member := range m {
		if results.Count() == limit {
			break
		}
		var passCount int
		for _, condFunc := range condFuncList {
			if condFunc(member) {
				passCount++
			}
		}
		if passCount == len(condFuncList) {
			results = append(results, member)
		}
	}
	return
}
