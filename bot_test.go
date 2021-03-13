package openwechat

import (
	"fmt"
	"testing"
)

func TestDefaultBot(t *testing.T) {
	bot := DefaultBot()
	messageHandler := func(message *Message) {
		fmt.Println(message.Content)
	}
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

func TestBotMessageSender(t *testing.T) {
	messageHandler := func(message *Message) {
		if message.IsSendByGroup() {
			sender, err := message.Sender()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(sender)
			if message.IsText() || message.Content == "test message" {
				message.ReplyText("hello")
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

func TestFriends_SearchByRemarkName(t *testing.T) {
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
	self, _ := bot.GetCurrentUser()
	friends, err := self.Friends()
	if err != nil {
		fmt.Println(err)
		return
	}
	d, err := friends[0].Detail()
	fmt.Println(d, err)
	firends2, err := friends.SearchByRemarkName("66")
	fmt.Println(firends2)
	fmt.Println(err)
}

func TestUser_GetAvatarResponse(t *testing.T) {
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
	self, _ := bot.GetCurrentUser()
	self.SaveAvatar("2.png")
	friend, err := self.Friends()
	if err != nil {
		fmt.Println(err)
		return
	}
	friend[0].SaveAvatar(friend[0].NickName + ".png")
}
