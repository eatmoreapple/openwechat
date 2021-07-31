package openwechat

import (
	"fmt"
	"testing"
)

func TestFormatEmoji(t *testing.T) {
	t.Log(FormatEmoji(`多吃点苹果<span class="emoji emoji1f34f"></span>高兴<span class="emoji emoji1f604"></span><span class="emoji emoji1f604"></span><span class="emoji emoji1f604"></span> 生气<span class="emoji emoji1f64e"></span> 点赞<span class="emoji emoji1f44d"></span>`))
}

func BenchmarkFormatEmojiString(b *testing.B) {
	str := `多吃点苹果<span class="emoji emoji1f34f"></span>高兴<span class="emoji emoji1f604"></span><span class="emoji emoji1f604"></span><span class="emoji emoji1f604"></span> 生气<span class="emoji emoji1f64e"></span> 点赞<span class="emoji emoji1f44d"></span>`
	b.SetBytes(int64(len(str)))
	// b.N会根据函数的运行时间取一个合适的值
	for i := 0; i < b.N; i++ {
		FormatEmoji(str)
	}
}

func BenchmarkFormatEmojiBlock(b *testing.B) {
	str := ""
	for ii := 0x1F301; ii <= 0x1F53D; ii++ {
		str += fmt.Sprintf(`<span class="emoji emoji%x"></span> `, ii)
	}
	b.SetBytes(int64(len(str)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FormatEmoji(str)
	}
}
