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
