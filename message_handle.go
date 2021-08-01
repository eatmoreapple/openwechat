package openwechat

// MessageHandler 消息处理函数
type MessageHandler func(msg *Message)

// MessageDispatcher 消息分发处理接口
// 跟 DispatchMessage 结合封装成 MessageHandler
type MessageDispatcher interface {
	Dispatch(msg *Message)
}

// DispatchMessage 跟 MessageDispatcher 结合封装成 MessageHandler
func DispatchMessage(dispatcher MessageDispatcher) func(msg *Message) {
	return func(msg *Message) { dispatcher.Dispatch(msg) }
}

// MessageDispatcher impl

// MessageContextHandler 消息处理函数
type MessageContextHandler func(ctx *MessageContext)

type MessageContextHandlerGroup []MessageContextHandler

// MessageContext 消息处理上下文对象
type MessageContext struct {
	index           int
	messageHandlers MessageContextHandlerGroup
	*Message
}

// Next 主动调用下一个消息处理函数(或开始调用)
func (c *MessageContext) Next() {
	c.index++
	for c.index <= len(c.messageHandlers) {
		handle := c.messageHandlers[c.index-1]
		handle(c)
		c.index++
	}
}

// 消息匹配函数,返回为true则表示匹配
type matchFunc func(*Message) bool

type matchNode struct {
	matchFunc matchFunc
	group     MessageContextHandlerGroup
}

type matchNodes []*matchNode

// MessageMatchDispatcher impl MessageDispatcher interface
//		dispatcher := NewMessageMatchDispatcher()
//		dispatcher.OnText(func(msg *Message){
//				msg.ReplyText("hello")
//		})
//		bot := DefaultBot()
//		bot.MessageHandler = DispatchMessage(dispatcher)
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
func (m *MessageMatchDispatcher) RegisterHandler(matchFunc matchFunc, handlers ...MessageContextHandler) {
	if matchFunc == nil {
		panic("matchFunc can not be nil")
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
