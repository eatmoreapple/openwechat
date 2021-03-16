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
	fileHelper, _ := self.FileHelper()
	fileHelper.SendText("6666")
	group, _ := self.Groups()
	friends, _ := self.Friends()
	fmt.Println(group.Search(1, func(group *Group) bool { return group.NickName == "厉害了" }))
	results := friends.Search(1, func(friend *Friend) bool { return friend.User.RemarkName == "阿青" }, func(friend *Friend) bool { return friend.Sex == 2 })
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
