package openwechat

import (
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
    return f.Self.SetRemarkNameToFriend(f, name)
}

// 发送自定义消息
func (f *Friend) SendMsg(msg *SendMessage) error {
    return f.Self.SendMessageToFriend(f, msg)
}

// 发送文本消息
func (f *Friend) SendText(content string) error {
    return f.Self.SendTextToFriend(f, content)
}

// 发送图片消息
func (f *Friend) SendImage(file *os.File) error {
    return f.Self.SendImageToFriend(f, file)
}

// 拉该好友入群
func (f *Friend) AddIntoGroup(groups ...*Group) error {
    return f.Self.AddFriendIntoManyGroups(f, groups...)
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

// 向slice的好友依次发送图片消息
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

// 发行消息给当前的群组
func (g *Group) SendMsg(msg *SendMessage) error {
    return g.Self.SendMessageToGroup(g, msg)
}

// 发行文本消息给当前的群组
func (g *Group) SendText(content string) error {
    return g.Self.SendTextToGroup(g, content)
}

// 发行图片消息给当前的群组
func (g *Group) SendImage(file *os.File) error {
    return g.Self.SendImageToGroup(g, file)
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
    return g.Self.AddFriendsIntoGroup(g, friends...)
}

// 从群聊中移除用户
// Deprecated
// 无论是网页版，还是程序上都不起作用
func (g *Group) RemoveMembers(members Members) error {
    return g.Self.RemoveMemberFromGroup(g, members)
}

type Groups []*Group

// 获取群组数量
func (g Groups) Count() int {
    return len(g)
}

// 获取第一个群组
func (g Groups) First() *Group {
    if g.Count() > 0 {
        return g[0]
    }
    return nil
}

// 获取最后一个群组
func (g Groups) Last() *Group {
    if g.Count() > 0 {
        return g[g.Count()-1]
    }
    return nil
}

// 向群组依次发送消息, 支持发送延迟
func (g Groups) SendMsg(msg *SendMessage, delay ...time.Duration) error {
    var total time.Duration
    for _, d := range delay {
        total += d
    }
    for _, group := range g {
        time.Sleep(total)
        if err := group.SendMsg(msg); err != nil {
            return err
        }
    }
    return nil
}

// 向群组依次发送文本消息, 支持发送延迟
func (g Groups) SendText(text string, delay ...time.Duration) error {
    var total time.Duration
    for _, d := range delay {
        total += d
    }
    for _, group := range g {
        time.Sleep(total)
        if err := group.SendText(text); err != nil {
            return err
        }
    }
    return nil
}

// 向群组依次发送图片消息, 支持发送延迟
func (g Groups) SendImage(file *os.File, delay ...time.Duration) error {
    var total time.Duration
    for _, d := range delay {
        total += d
    }
    for _, group := range g {
        time.Sleep(total)
        if err := group.SendImage(file); err != nil {
            return err
        }
    }
    return nil
}

// 根据用户名查找群组
func (g Groups) SearchByUserName(limit int, username string) (results Groups) {
    return g.Search(limit, func(group *Group) bool { return group.UserName == username })
}

// 根据昵称查找群组
func (g Groups) SearchByNickName(limit int, nickName string) (results Groups) {
    return g.Search(limit, func(group *Group) bool { return group.NickName == nickName })
}

// 根据备注查找群组
func (g Groups) SearchByRemarkName(limit int, remarkName string) (results Groups) {
    return g.Search(limit, func(group *Group) bool { return group.RemarkName == remarkName })
}

// 根据自定义条件查找群组
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

// 公众号对象
type Mp struct{ *User }

func (m Mp) String() string {
    return fmt.Sprintf("<Mp:%s>", m.NickName)
}

// 公众号组对象
type Mps []*Mp

// 数量统计
func (m Mps) Count() int {
    return len(m)
}

// 获取第一个
func (m Mps) First() *Mp {
    if m.Count() > 0 {
        return m[0]
    }
    return nil
}

// 获取最后一个
func (m Mps) Last() *Mp {
    if m.Count() > 0 {
        return m[m.Count()-1]
    }
    return nil
}

// 根据自定义条件查找
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

// 判断是否为好友
func isFriend(user User) bool {
    return !isGroup(user) && strings.HasPrefix(user.UserName, "@") && user.VerifyFlag == 0
}

// 判断是否为群组
func isGroup(user User) bool {
    return strings.HasPrefix(user.UserName, "@@") && user.VerifyFlag == 0
}

// 判断是否为公众号
func isMP(user User) bool {
    return user.VerifyFlag == 8 || user.VerifyFlag == 24 || user.VerifyFlag == 136
}
