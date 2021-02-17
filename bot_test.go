package openwechat

import (
	"fmt"
	"testing"
)

func TestDefaultBot(t *testing.T) {
	messageHandler := func(message Message) {
		fmt.Println(message)
	}
	bot := DefaultBot(messageHandler)
	bot.UUIDCallback = PrintlnQrcodeUrl
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}
	//for bot.Alive() {
	//	message := messageHandler.GetMessage()
	//	if message.Content == "6666" {
	//		err := message.ReplyText("nihao")
	//		fmt.Println(err)
	//	}
	//	fmt.Println(message)
	//}
}
