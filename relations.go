package openwechat

import (
	"fmt"
	"os"
	"reflect"
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

func (f Friends) Search(cond Cond) (friends Friends, found bool) {
	for _, member := range f {
		value := reflect.ValueOf(member).Elem()
		for k, v := range cond {
			if field := value.FieldByName(k); field.IsValid() {
				if field.Interface() == v {
					found = true
					if friends == nil {
						friends = make(Friends, 0)
					}
					friends = append(friends, member)
				}
			}
		}
	}
	return
}

func (f Friends) SendMsg(msg *SendMessage) error {
	for _, friend := range f {
		if err := friend.SendMsg(msg); err != nil {
			return err
		}
	}
	return nil
}

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

func (g Groups) Search(cond Cond) (groups Groups, found bool) {
	for _, member := range g {
		value := reflect.ValueOf(member).Elem()
		for k, v := range cond {
			if field := value.FieldByName(k); field.IsValid() {
				if field.Interface() == v {
					found = true
					if groups == nil {
						groups = make(Groups, 0)
					}
					groups = append(groups, member)
				}
			}
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
