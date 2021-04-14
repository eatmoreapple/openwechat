package openwechat

import "testing"

func TestFormatEmoji(t *testing.T) {
	t.Log(FormatEmoji(`多吃点苹果<span class="emoji emoji1f34f"></span>`))
}
