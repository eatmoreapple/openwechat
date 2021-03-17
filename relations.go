package openwechat

import (
	"fmt"
	"os"
	"strings"
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
func (f Friends) SendMsg(msg *SendMessage) error {
	for _, friend := range f {
		if err := friend.SendMsg(msg); err != nil {
			return err
		}
	}
	return nil
}

// 向slice的好友依次发送文本消息
func (f Friends) SendText(text string) error {
	for _, friend := range f {
		if err := friend.SendText(text); err != nil {
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
