package openwechat

import (
	"errors"
	"strings"
)

// MessageHandler 消息处理函数
type MessageHandler func(msg *Message)

// MessageDispatcher 消息分发处理接口
// 跟 DispatchMessage 结合封装成 MessageHandler
type MessageDispatcher interface {
	Dispatch(msg *Message)
}

// MessageContextHandler 消息处理函数
type MessageContextHandler func(ctx *MessageContext)

type MessageContextHandlerGroup []MessageContextHandler

// MessageContext 消息处理上下文对象
type MessageContext struct {
	index           int
	abortIndex      int
	messageHandlers MessageContextHandlerGroup
	*Message
}

// Next 主动调用下一个消息处理函数(或开始调用)
func (c *MessageContext) Next() {
	c.index++
	for c.index <= len(c.messageHandlers) {
		if c.IsAbort() {
			return
		}
		handle := c.messageHandlers[c.index-1]
		handle(c)
		c.index++
	}
}

// IsAbort 判断是否被中断
func (c *MessageContext) IsAbort() bool {
	return c.abortIndex > 0
}

// Abort 中断当前消息处理, 不会调用下一个消息处理函数, 但是不会中断当前的处理函数
func (c *MessageContext) Abort() {
	c.abortIndex = c.index
}

// AbortHandler 获取当前中断的消息处理函数
func (c *MessageContext) AbortHandler() MessageContextHandler {
	if c.abortIndex > 0 {
		return c.messageHandlers[c.abortIndex-1]
	}
	return nil
}

// MatchFunc 消息匹配函数,返回为true则表示匹配
type MatchFunc func(*Message) bool

// MatchFuncList 将多个MatchFunc封装成一个MatchFunc
func MatchFuncList(matchFuncs ...MatchFunc) MatchFunc {
	return func(message *Message) bool {
		for _, matchFunc := range matchFuncs {
			if !matchFunc(message) {
				return false
			}
		}
		return true
	}
}

type matchNode struct {
	matchFunc MatchFunc
	group     MessageContextHandlerGroup
}

type matchNodes []*matchNode

// MessageMatchDispatcher impl MessageDispatcher interface
//
//	dispatcher := NewMessageMatchDispatcher()
//	dispatcher.OnText(func(msg *Message){
//			msg.ReplyText("hello")
//	})
//	bot := DefaultBot()
//	bot.MessageHandler = DispatchMessage(dispatcher)
type MessageMatchDispatcher struct {
	async      bool
	matchNodes matchNodes
}

// NewMessageMatchDispatcher Constructor
func NewMessageMatchDispatcher() *MessageMatchDispatcher {
	return &MessageMatchDispatcher{}
}

// SetAsync 设置是否异步处理
func (m *MessageMatchDispatcher) SetAsync(async bool) {
	m.async = async
}

// Dispatch impl MessageDispatcher
// 遍历 MessageMatchDispatcher 所有的消息处理函数
// 获取所有匹配上的函数
// 执行处理的消息处理方法
func (m *MessageMatchDispatcher) Dispatch(msg *Message) {
	var group MessageContextHandlerGroup
	for _, node := range m.matchNodes {
		if node.matchFunc(msg) {
			group = append(group, node.group...)
		}
	}
	ctx := &MessageContext{Message: msg, messageHandlers: group}
	if m.async {
		go m.do(ctx)
	} else {
		m.do(ctx)
	}
}

func (m *MessageMatchDispatcher) do(ctx *MessageContext) {
	ctx.Next()
}

// RegisterHandler 注册消息处理函数, 根据自己的需求自定义
// matchFunc返回true则表示处理对应的handlers
func (m *MessageMatchDispatcher) RegisterHandler(matchFunc MatchFunc, handlers ...MessageContextHandler) {
	if matchFunc == nil {
		panic("MatchFunc can not be nil")
	}
	node := &matchNode{matchFunc: matchFunc, group: handlers}
	m.matchNodes = append(m.matchNodes, node)
}

// OnText 注册处理消息类型为Text的处理函数
func (m *MessageMatchDispatcher) OnText(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsText() }, handlers...)
}

// OnImage 注册处理消息类型为Image的处理函数
func (m *MessageMatchDispatcher) OnImage(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsPicture() }, handlers...)
}

// OnEmoticon 注册处理消息类型为Emoticon的处理函数(表情包)
func (m *MessageMatchDispatcher) OnEmoticon(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsEmoticon() }, handlers...)
}

// OnVoice 注册处理消息类型为Voice的处理函数
func (m *MessageMatchDispatcher) OnVoice(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsVoice() }, handlers...)
}

// OnFriendAdd 注册处理消息类型为FriendAdd的处理函数
func (m *MessageMatchDispatcher) OnFriendAdd(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsFriendAdd() }, handlers...)
}

// OnCard 注册处理消息类型为Card的处理函数
func (m *MessageMatchDispatcher) OnCard(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsCard() }, handlers...)
}

// OnMedia 注册处理消息类型为Media(多媒体消息，包括但不限于APP分享、文件分享)的处理函数
func (m *MessageMatchDispatcher) OnMedia(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsMedia() }, handlers...)
}

// OnFriendByNickName 注册根据好友昵称是否匹配的消息处理函数
func (m *MessageMatchDispatcher) OnFriendByNickName(nickName string, handlers ...MessageContextHandler) {
	matchFunc := func(message *Message) bool {
		if message.IsSendByFriend() {
			sender, err := message.Sender()
			return err == nil && sender.NickName == nickName
		}
		return false
	}
	m.RegisterHandler(matchFunc, handlers...)
}

// OnFriend 注册发送者为好友的处理函数
func (m *MessageMatchDispatcher) OnFriend(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsSendByFriend() }, handlers...)
}

// OnGroup 注册发送者为群组的处理函数
func (m *MessageMatchDispatcher) OnGroup(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsSendByGroup() }, handlers...)
}

// OnUser 注册根据消息发送者的行为是否匹配的消息处理函数
func (m *MessageMatchDispatcher) OnUser(f func(user *User) bool, handlers ...MessageContextHandler) {
	mf := func(message *Message) bool {
		sender, err := message.Sender()
		if err != nil {
			return false
		}
		return f(sender)
	}
	m.RegisterHandler(mf, handlers...)
}

// OnFriendByRemarkName 注册根据好友备注是否匹配的消息处理函数
func (m *MessageMatchDispatcher) OnFriendByRemarkName(remarkName string, handlers ...MessageContextHandler) {
	f := func(user *User) bool {
		return user.IsFriend() && user.RemarkName == remarkName
	}
	m.OnUser(f, handlers...)
}

// OnGroupByGroupName 注册根据群名是否匹配的消息处理函数
func (m *MessageMatchDispatcher) OnGroupByGroupName(groupName string, handlers ...MessageContextHandler) {
	f := func(user *User) bool {
		return user.IsGroup() && user.NickName == groupName
	}
	m.OnUser(f, handlers...)
}

// OnTrickled 注册处理消息类型为拍一拍的处理函数
func (m *MessageMatchDispatcher) OnTrickled(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsTickled() }, handlers...)
}

// OnRecalled 注册撤回消息类型的处理函数
func (m *MessageMatchDispatcher) OnRecalled(handlers ...MessageContextHandler) {
	m.RegisterHandler(func(message *Message) bool { return message.IsRecalled() }, handlers...)
}

// AsMessageHandler 将MessageMatchDispatcher转换为MessageHandler
func (m *MessageMatchDispatcher) AsMessageHandler() MessageHandler {
	return func(msg *Message) {
		m.Dispatch(msg)
	}
}

type MessageSenderMatchFunc func(user *User) bool

// SenderMatchFunc 抽象的匹配发送者特征的处理函数
//
//	    dispatcher := NewMessageMatchDispatcher()
//		   matchFuncList := MatchFuncList(SenderFriendRequired(), SenderNickNameContainsMatchFunc("多吃点苹果"))
//		   dispatcher.RegisterHandler(matchFuncList, func(ctx *MessageContext) {
//			     do your own business
//		   })
func SenderMatchFunc(matchFuncs ...MessageSenderMatchFunc) MatchFunc {
	return func(message *Message) bool {
		sender, err := message.Sender()
		if err != nil {
			return false
		}
		for _, matchFunc := range matchFuncs {
			if !matchFunc(sender) {
				return false
			}
		}
		return true
	}
}

// SenderFriendRequired 只匹配好友
func SenderFriendRequired() MatchFunc {
	return SenderMatchFunc(func(user *User) bool { return user.IsFriend() })
}

// SenderGroupRequired 只匹配群组
func SenderGroupRequired() MatchFunc {
	return SenderMatchFunc(func(user *User) bool { return user.IsGroup() })
}

// SenderMpRequired 只匹配公众号
func SenderMpRequired() MatchFunc {
	return SenderMatchFunc(func(user *User) bool { return user.IsMP() })
}

// SenderNickNameEqualMatchFunc 根据用户昵称是否等于指定字符串的匹配函数
func SenderNickNameEqualMatchFunc(nickname string) MatchFunc {
	return SenderMatchFunc(func(user *User) bool { return user.NickName == nickname })
}

// SenderRemarkNameEqualMatchFunc 根据用户备注是否等于指定字符串的匹配函数
func SenderRemarkNameEqualMatchFunc(remarkName string) MatchFunc {
	return SenderMatchFunc(func(user *User) bool { return user.RemarkName == remarkName })
}

// SenderNickNameContainsMatchFunc 根据用户昵称是否包含指定字符串的匹配函数
func SenderNickNameContainsMatchFunc(nickname string) MatchFunc {
	return SenderMatchFunc(func(user *User) bool { return strings.Contains(user.NickName, nickname) })
}

// SenderRemakeNameContainsFunc  根据用户备注名是否包含指定字符串的匹配函数
func SenderRemakeNameContainsFunc(remakeName string) MatchFunc {
	return SenderMatchFunc(func(user *User) bool { return strings.Contains(user.RemarkName, remakeName) })
}

// MessageErrorHandler 获取消息时发生了错误的处理函数
// 参数err为获取消息时发生的错误，返回值为处理后的错误
// 如果返回nil，则表示忽略该错误，否则将继续传递该错误
type MessageErrorHandler func(err error) error

// defaultMessageErrorHandler 默认的SyncCheck错误处理函数
func defaultMessageErrorHandler(err error) error {
	var ret Ret
	if errors.As(err, &ret) {
		switch ret {
		case failedLoginCheck, cookieInvalid, failedLoginWarn:
			return ret
		}
	}
	return nil
}
