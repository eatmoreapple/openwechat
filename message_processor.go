package openwechat

import (
	"encoding/json"
	"html"
	"strings"
)

type MessageProcessor interface {
	ProcessMessage(msg *Message)
}

type MessageProcessorGroup []MessageProcessor

func (g MessageProcessorGroup) ProcessMessage(msg *Message) {
	for _, processor := range g {
		processor.ProcessMessage(msg)
	}
}

// 保存消息原始内容
type messageRowContentProcessor struct{}

func (m *messageRowContentProcessor) ProcessMessage(msg *Message) {
	raw, _ := json.Marshal(msg)
	msg.Raw = raw
	msg.RawContent = msg.Content
}

// 保存发送者在群里的用户名
type senderInGroupMessageProcessor struct{}

func (s *senderInGroupMessageProcessor) ProcessMessage(msg *Message) {
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
type atMessageProcessor struct{}

func (g *atMessageProcessor) ProcessMessage(msg *Message) {
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
type wrapLineMessageProcessor struct{}

func (w *wrapLineMessageProcessor) ProcessMessage(msg *Message) {
	msg.Content = strings.Replace(msg.Content, `<br/>`, "\n", -1)
}

// 处理消息中的html转义字符
type unescapeHTMLMessageProcessor struct{}

func (u *unescapeHTMLMessageProcessor) ProcessMessage(msg *Message) {
	msg.Content = html.UnescapeString(msg.Content)
}

// 处理消息中的emoji表情
type emojiMessageProcessor struct{}

func (e *emojiMessageProcessor) ProcessMessage(msg *Message) {
	msg.Content = FormatEmoji(msg.Content)
}

// 尝试获取群聊中的消息的发送者
type tryToFindGroupMessageProcessor struct{}

func (t *tryToFindGroupMessageProcessor) ProcessMessage(msg *Message) {
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
	defaultMessageProcessor MessageProcessor = MessageProcessorGroup{
		&messageRowContentProcessor{},
		&senderInGroupMessageProcessor{},
		&atMessageProcessor{},
		&wrapLineMessageProcessor{},
		&unescapeHTMLMessageProcessor{},
		&emojiMessageProcessor{},
		&tryToFindGroupMessageProcessor{},
	}
)
