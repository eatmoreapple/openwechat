package openwechat

import (
	"fmt"
	"testing"
)

func TestDefaultBot(t *testing.T) {
	messageHandler := func(message *Message) {
		fmt.Println(message)
	}
	bot := DefaultBot()
	bot.RegisterMessageHandler(messageHandler)
	bot.UUIDCallback = PrintlnQrcodeUrl
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}
	bot.Block()
}

func TestBotMessageHandler(t *testing.T) {
	messageHandler := func(message *Message) {
		if message.IsSendByGroup() {
			sender, err := message.Sender()
			if err != nil {
				fmt.Println(err)
				return
			}
			group := Group{sender}
			members, err := group.Members()
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, member := range members {
				fmt.Println(member)
			}
			if message.IsText() {
				message.ReplyText(message.Content)
			}
		}
	}
	bot := DefaultBot()
	bot.RegisterMessageHandler(messageHandler)
	bot.UUIDCallback = PrintlnQrcodeUrl
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}
	bot.Block()
}
