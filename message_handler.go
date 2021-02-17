package openwechat

type MessageHandler func(message *Message)

type MessageHandlerGroup struct {
	handlers []MessageHandler
}

func (m MessageHandlerGroup) ProcessMessage(message *Message) {
	for _, handler := range m.handlers {
		handler(message)
	}
}

func (m *MessageHandlerGroup) RegisterHandler(handler MessageHandler) {
	m.handlers = append(m.handlers, handler)
}
