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

func (f *Friend) RemarkName(name string) error {
	return f.remakeName(name)
}

func (f *Friend) SendMsg(msg *SendMessage) error {
	return f.sendMsg(msg)
}

func (f *Friend) SendText(content string) error {
	return f.sendText(content)
}

func (f *Friend) SendImage(file *os.File) error {
	return f.sendImage(file)
}

type Friends []*Friend

func (f Friends) Count() int {
	return len(f)
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

func isFriend(user User) bool {
	return !isGroup(user) && strings.HasPrefix(user.UserName, "@") && user.VerifyFlag == 0
}

func isGroup(user User) bool {
	return strings.HasPrefix(user.UserName, "@@") && user.VerifyFlag == 0
}
