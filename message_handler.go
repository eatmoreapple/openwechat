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
	if m.handlers == nil {
		m.handlers = make([]MessageHandler, 0)
	}
	m.handlers = append(m.handlers, handler)
}
