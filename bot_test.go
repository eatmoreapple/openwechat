package openwechat

import (
	"fmt"
	"testing"
)

func TestDefaultBot(t *testing.T) {
	bot := DefaultBot()
	bot.MessageHandler = func(message *Message) {
		if message.Content == "logout" {
			bot.Logout()
		}
		fmt.Println(message.Content)
	}
	bot.UUIDCallback = PrintlnQrcodeUrl
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}
	self, _ := bot.GetCurrentUser()
	self.SaveAvatar("2.png")
	fileHelper, _ := self.FileHelper()
	fileHelper.SendText("6666")
	group, _ := self.Groups()
	friends, _ := self.Friends()
	fmt.Println(group.Search(1, func(group *Group) bool { return group.NickName == "厉害了" }))
	results := friends.Search(1, func(friend *Friend) bool { return friend.RemarkName == "阿青" }, func(friend *Friend) bool { return friend.Sex == 2 })
	fmt.Println(results)
	fmt.Println(bot.Block())
}

func TestUser_GetAvatarResponse(t *testing.T) {

	bot := DefaultBot()
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

func getDefaultLoginBot() *Bot {
	bot := DefaultBot()
	bot.UUIDCallback = PrintlnQrcodeUrl
	return bot
}

func TestSendFile(t *testing.T) {
	bot := getDefaultLoginBot()
	bot.MessageHandler = func(msg *Message) {
		user, err := msg.Sender()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(msg.Content)
		fmt.Println(user.NickName)
	}
	if err := bot.Login(); err != nil {
		t.Error(err)
		return
	}
	self, err := bot.GetCurrentUser()
	if err != nil {
		t.Error(err)
		return
	}
	fileHelper, _ := self.FileHelper()
	fileHelper.sendText("666")
	bot.Block()
}
