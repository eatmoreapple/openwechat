package openwechat

import (
	"encoding/json"
	"html"
	"strings"
)

type MessageObserver interface {
	OnMessageReceive(msg *Message)
}

type MessageObserverGroup []MessageObserver

func (g MessageObserverGroup) OnMessageReceive(msg *Message) {
	for _, observer := range g {
		observer.OnMessageReceive(msg)
	}
}

// 保存消息原始内容
type messageRowContentObserver struct{}

func (m *messageRowContentObserver) OnMessageReceive(msg *Message) {
	raw, _ := json.Marshal(msg)
	msg.Raw = raw
	msg.RawContent = msg.Content
}

// 保存发送者在群里的用户名
type senderInGroupObserver struct{}

func (s *senderInGroupObserver) OnMessageReceive(msg *Message) {
	if !msg.IsSendByGroup() || msg.IsSystem() || msg.IsSendBySelf() {
		return
	}
	data := strings.Split(msg.Content, ":<br/>")
	if len(data) < 2 {
		return
	}
	msg.Content = strings.Join(data[1:], "")
	msg.senderUserNameInGroup = data[0]
}

// 检查消息是否被@了, 不是特别严谨
type atMessageObserver struct{}

func (g *atMessageObserver) OnMessageReceive(msg *Message) {
	if !msg.IsSendByGroup() {
		return
	}
	if msg.IsSystem() {
		return
	}
	if msg.IsSendBySelf() {
		// 这块不严谨，但是只能这么干了
		msg.isAt = strings.Contains(msg.Content, "@") || strings.Contains(msg.Content, "\u2005")
		return
	}
	if strings.Contains(msg.Content, "@") {
		sender, err := msg.Sender()
		if err == nil {
			receiver := sender.MemberList.SearchByUserName(1, msg.ToUserName)
			if receiver != nil {
				displayName := receiver.First().DisplayName
				if displayName == "" {
					displayName = receiver.First().NickName
				}
				var atFlag string
				msgContent := FormatEmoji(msg.Content)
				atName := FormatEmoji(displayName)
				if strings.Contains(msgContent, "\u2005") {
					atFlag = "@" + atName + "\u2005"
				} else {
					atFlag = "@" + atName
				}
				msg.isAt = strings.Contains(msgContent, atFlag) || strings.HasSuffix(msgContent, atFlag)
			}
		}
	}
}

// 处理消息中的换行符
type wrapLineMessageObserver struct{}

func (w *wrapLineMessageObserver) OnMessageReceive(msg *Message) {
	msg.Content = strings.Replace(msg.Content, `<br/>`, "\n", -1)
}

// 处理消息中的html转义字符
type unescapeHTMLMessageObserver struct{}

func (u *unescapeHTMLMessageObserver) OnMessageReceive(msg *Message) {
	msg.Content = html.UnescapeString(msg.Content)
}

// 处理消息中的emoji表情
type emojiMessageObserver struct{}

func (e *emojiMessageObserver) OnMessageReceive(msg *Message) {
	msg.Content = FormatEmoji(msg.Content)
}

// 尝试获取群聊中的消息的发送者
type tryToFindGroupMemberObserver struct{}

func (t *tryToFindGroupMemberObserver) OnMessageReceive(msg *Message) {
	if msg.IsSendByGroup() {
		if msg.FromUserName == msg.Owner().UserName {
			return
		}
		// 首先尝试从缓存里面查找, 如果没有找到则从服务器获取
		members, err := msg.Owner().Members()
		if err != nil {
			return
		}
		_, exist := members.GetByUserName(msg.FromUserName)
		if !exist {
			owner := msg.Owner()
			// 找不到, 从服务器获取
			user := newUser(owner, msg.FromUserName)
			_ = user.Detail()
			owner.members = owner.members.Append(user)
			owner.groups = owner.members.Groups()
		}
	}
}

var (
	defaultMessageObserver MessageObserver = MessageObserverGroup{
		&messageRowContentObserver{},
		&senderInGroupObserver{},
		&atMessageObserver{},
		&wrapLineMessageObserver{},
		&unescapeHTMLMessageObserver{},
		&emojiMessageObserver{},
		&tryToFindGroupMemberObserver{},
	}
)
