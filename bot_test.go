package openwechat

import "testing"

func defaultBot() *Bot {
	bot := DefaultBot()
	bot.UUIDCallback = PrintlnQrcodeUrl
	return bot
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
