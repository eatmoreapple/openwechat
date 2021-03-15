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

func (f Friends) First() *Friend {
	if f.Count() > 0 {
		return f[0]
	}
	return nil
}

func (f Friends) Last() *Friend {
	if f.Count() > 0 {
		return f[f.Count()-1]
	}
	return nil
}

func (f Friends) SearchByUserName(username string, limit int) (results Friends, found bool) {
	if limit <= 0 {
		limit = f.Count()
	}
	for _, friend := range f {
		if results.Count() == limit {
			break
		}
		if friend.UserName == username {
			found = true
			results = append(results, friend)
		}
	}
	return
}

func (f Friends) SearchByNickName(nickName string, limit int) (results Friends, found bool) {
	if limit <= 0 {
		limit = f.Count()
	}
	for _, friend := range f {
		if results.Count() == limit {
			break
		}
		if friend.NickName == nickName {
			found = true
			results = append(results, friend)
		}
	}
	return
}

func (f Friends) SearchByRemarkName(remarkName string, limit int) (results Friends, found bool) {
	if limit <= 0 {
		limit = f.Count()
	}
	for _, friend := range f {
		if results.Count() == limit {
			break
		}
		if friend.User.RemarkName == remarkName {
			found = true
			results = append(results, friend)
		}
	}
	return
}

func (f Friends) Search(cond Cond, limit int) (results Friends, found bool) {
	if len(cond) == 1 {
		for k, v := range cond {
			switch k {
			case "UserName":
				if value, ok := v.(string); ok {
					return f.SearchByUserName(value, limit)
				}
			case "NickName":
				if value, ok := v.(string); ok {
					return f.SearchByNickName(value, limit)
				}
			case "RemarkName":
				if value, ok := v.(string); ok {
					return f.SearchByRemarkName(value, limit)
				}
			}
		}
	}
	if limit <= 0 {
		limit = f.Count()
	}
	for _, friend := range f {
		if results.Count() == limit {
			break
		}
		value := reflect.ValueOf(friend).Elem()
		var matchCount int
		for k, v := range cond {
			if field := value.FieldByName(k); field.IsValid() {
				if field.Interface() != v {
					break
				}
				matchCount++
			}
		}
		if matchCount == len(cond) {
			found = true
			results = append(results, friend)
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

func (g Groups) SearchByUserName(username string, limit int) (results Groups, found bool) {
	if limit <= 0 {
		limit = g.Count()
	}
	for _, group := range g {
		if results.Count() == limit {
			break
		}
		if group.UserName == username {
			found = true
			results = append(results, group)
		}
	}
	return
}

func (g Groups) SearchByNickName(nickName string, limit int) (results Groups, found bool) {
	if limit <= 0 {
		limit = g.Count()
	}
	for _, group := range g {
		if results.Count() == limit {
			break
		}
		if group.NickName == nickName {
			found = true
			results = append(results, group)
		}
	}
	return
}

func (g Groups) SearchByRemarkName(remarkName string, limit int) (results Groups, found bool) {
	if limit <= 0 {
		limit = g.Count()
	}
	for _, group := range g {
		if results.Count() == limit {
			break
		}
		if group.User.RemarkName == remarkName {
			found = true
			results = append(results, group)
		}
	}
	return
}

func (g Groups) Search(cond Cond, limit int) (results Groups, found bool) {

	if len(cond) == 1 {
		for k, v := range cond {
			switch k {
			case "UserName":
				if value, ok := v.(string); ok {
					return g.SearchByUserName(value, limit)
				}
			case "NickName":
				if value, ok := v.(string); ok {
					return g.SearchByNickName(value, limit)
				}
			case "RemarkName":
				if value, ok := v.(string); ok {
					return g.SearchByRemarkName(value, limit)
				}
			}
		}
	}
	if limit <= 0 {
		limit = g.Count()
	}
	for _, group := range g {
		if g.Count() == limit {
			break
		}
		value := reflect.ValueOf(group).Elem()
		var matchCount int
		for k, v := range cond {
			if field := value.FieldByName(k); field.IsValid() {
				if field.Interface() != v {
					break
				}
				matchCount++
			}
		}
		if matchCount == len(cond) {
			found = true
			results = append(results, group)
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
