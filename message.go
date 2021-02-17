package openwechat

import (
	"errors"
	"os"
	"strings"
	"time"
)

type Message struct {
	AppInfo struct {
		AppID string
		Type  int
	}
	AppMsgType            int
	Content               string
	CreateTime            int64
	EncryFileName         string
	FileName              string
	FileSize              string
	ForwardFlag           int
	FromUserName          string
	HasProductId          int
	ImgHeight             int
	ImgStatus             int
	ImgWidth              int
	MediaId               string
	MsgId                 string
	MsgType               int
	NewMsgId              int64
	OriContent            string
	PlayLength            int64
	RecommendInfo         RecommendInfo
	Status                int
	StatusNotifyCode      int
	StatusNotifyUserName  string
	SubMsgType            int
	Ticket                string
	ToUserName            string
	Url                   string
	VoiceLength           int
	Bot                   *Bot
	senderInGroupUserName string
}

func (m *Message) Sender() (*User, error) {
	members, err := m.Bot.self.Members(true)
	if err != nil {
		return nil, err
	}
	if m.FromUserName == m.Bot.self.User.UserName {
		return m.Bot.self.User, nil
	}
	for _, member := range members {
		if member.UserName == m.FromUserName {
			return member.Detail()
		}
	}
	return nil, errors.New("no such user found")
}

func (m *Message) SenderInGroup() (*User, error) {
	if !m.IsSendByGroup() {
		return nil, errors.New("message is not from group")
	}
	group, err := m.Sender()
	if err != nil {
		return nil, err
	}
	group, err = group.Detail()
	if err != nil {
		return nil, err
	}
	for _, member := range group.MemberList {
		if m.senderInGroupUserName == member.UserName {
			return member, nil
		}
	}
	return nil, errors.New("no such user found")
}

//
func (m *Message) IsSendBySelf() bool {
	return m.FromUserName == m.Bot.self.User.UserName
}

func (m *Message) IsSendByFriend() bool {
	return !m.IsSendByGroup() && strings.HasPrefix(m.FromUserName, "@")
}

func (m *Message) IsSendByGroup() bool {
	return strings.HasPrefix(m.FromUserName, "@@")
}

func (m *Message) Reply(msgType int, content, mediaId string) error {
	msg := NewSendMessage(msgType, content, m.Bot.self.User.UserName, m.FromUserName, mediaId)
	info := m.Bot.storage.GetLoginInfo()
	request := m.Bot.storage.GetBaseRequest()
	return m.Bot.Caller.WebWxSendMsg(msg, info, request)
}

func (m *Message) ReplyText(content string) error {
	return m.Reply(TextMessage, content, "")
}

func (m *Message) ReplyImage(file *os.File) error {
	info := m.Bot.storage.GetLoginInfo()
	request := m.Bot.storage.GetBaseRequest()
	return m.Bot.Caller.WebWxSendImageMsg(file, request, info, m.Bot.self.UserName, m.FromUserName)
}

func (m *Message) IsText() bool {
	return m.MsgType == 1 && m.Url == ""
}

func (m *Message) IsMap() bool {
	return m.MsgType == 1 && m.Url != ""
}

func (m *Message) IsPicture() bool {
	return m.MsgType == 3 || m.MsgType == 47
}

func (m *Message) IsVoice() bool {
	return m.MsgType == 34
}

func (m *Message) IsFriendAdd() bool {
	return m.MsgType == 37
}

func (m *Message) IsCard() bool {
	return m.MsgType == 42
}

func (m *Message) IsVideo() bool {
	return m.MsgType == 43 || m.MsgType == 62
}

func (m *Message) IsSharing() bool {
	return m.MsgType == 49
}

func (m *Message) IsRecalled() bool {
	return m.MsgType == 10002
}

func (m *Message) IsSystem() bool {
	return m.MsgType == 10000
}

//func (m Message) Agree() error {
//	if !m.IsFriendAdd() {
//		return fmt.Errorf("the excepted message type is 37, but got %d", m.MsgType)
//	}
//	m.ClientManager.Client.WebWxVerifyUser(m.ClientManager.storage, m.RecommendInfo, "")
//}

type SendMessage struct {
	Type         int
	Content      string
	FromUserName string
	ToUserName   string
	LocalID      int64
	ClientMsgId  int64
	MediaId      string
}

func NewSendMessage(msgType int, content, fromUserName, toUserName, mediaId string) *SendMessage {
	return &SendMessage{
		Type:         msgType,
		Content:      content,
		FromUserName: fromUserName,
		ToUserName:   toUserName,
		LocalID:      time.Now().Unix() * 1e4,
		ClientMsgId:  time.Now().Unix() * 1e4,
		MediaId:      mediaId,
	}
}

func NewTextSendMessage(content, fromUserName, toUserName string) *SendMessage {
	return NewSendMessage(TextMessage, content, fromUserName, toUserName, "")
}

func NewMediaSendMessage(msgType int, fromUserName, toUserName, mediaId string) *SendMessage {
	return NewSendMessage(msgType, "", fromUserName, toUserName, mediaId)
}

type RecommendInfo struct {
	Alias      string
	AttrStatus int64
	City       string
	Content    string
	NickName   string
	OpCode     int
	Province   string
	QQNum      int64
	Scene      int
	Sex        int
	Signature  string
	Ticket     string
	UserName   string
	VerifyFlag int
}

func processMessage(message *Message, bot *Bot) {
	message.Bot = bot
	if message.IsSendByGroup() {
		data := strings.Split(message.Content, ":<br/>")
		message.Content = strings.Join(data[1:], "")
		message.senderInGroupUserName = data[0]
	}
}
