package openwechat

import "testing"

func TestFormatEmoji(t *testing.T) {
	t.Log(FormatEmoji(`多吃点苹果<span class="emoji emoji1f34f"></span>`))
}

func TestSendEmoji(t *testing.T) {
	self, err := getSelf()
	if err != nil {
		t.Error(err)
		return
	}
	f, err := self.FileHelper()
	if err != nil {
		t.Error(err)
		return
	}
	_ = f.SendText(Emoji.Dagger)
}
