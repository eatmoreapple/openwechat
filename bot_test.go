package openwechat

import "testing"

func defaultBot(modes ...mode) *Bot {
	bot := DefaultBot(modes...)
	bot.UUIDCallback = PrintlnQrcodeUrl
	return bot
}

func getSelf(modes ...mode) (*Self, error) {
	bot := defaultBot(modes...)
	if err := bot.Login(); err != nil {
		return nil, err
	}
	return bot.GetCurrentUser()
}

func TestBotLogin(t *testing.T) {
	bot := defaultBot()
	if err := bot.Login(); err != nil {
		t.Error(err)
		return
	}
	self, err := bot.GetCurrentUser()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(self.NickName)
}

func TestFriend(t *testing.T) {
	self, err := getSelf()
	if err != nil {
		t.Error(err)
		return
	}
	friends, err := self.Friends()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(friends)
}
