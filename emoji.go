package openwechat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	emojiRegexp = regexp.MustCompile(`<span class="emoji emoji(.*?)"></span>`)
)

// Emoji 表情
// 字段太多了,污染命名空间,封装成struct返回
var Emoji = struct {
	Smile        string
	Grimace      string
	Drool        string
	Scowl        string
	CoolGuy      string
	Sob          string
	Shy          string
	Silent       string
	Sleep        string
	Cry          string
	Awkward      string
	Angry        string
	Tongue       string
	Grin         string
	Surprise     string
	Frown        string
	Ruthless     string
	Blush        string
	Scream       string
	Puke         string
	Chuckle      string
	Joyful       string
	Slight       string
	Smug         string
	Hungry       string
	Drowsy       string
	Panic        string
	Sweat        string
	Laugh        string
	Commando     string
	Determined   string
	Scold        string
	Shocked      string
	Shhh         string
	Dizzy        string
	Tormented    string
	Toasted      string
	Skull        string
	Hammer       string
	Wave         string
	Speechless   string
	NosePick     string
	Clap         string
	Shame        string
	Trick        string
	BahL         string
	BahR         string
	Yawn         string
	PoohPooh     string
	Shrunken     string
	TearingUp    string
	Sly          string
	Kiss         string
	Wrath        string
	Whimper      string
	Cleaver      string
	Watermelon   string
	Beer         string
	Basketball   string
	PingPong     string
	Coffee       string
	Rice         string
	Pig          string
	Rose         string
	Wilt         string
	Lips         string
	Heart        string
	BrokenHeart  string
	Cake         string
	Lightning    string
	Bomb         string
	Dagger       string
	Soccer       string
	Ladybug      string
	Poop         string
	Moon         string
	Sun          string
	Gift         string
	Hug          string
	ThumbsUp     string
	ThumbsDown   string
	Shake        string
	Peace        string
	Fight        string
	Beckon       string
	Fist         string
	Pinky        string
	RockOn       string
	Nuhuh        string
	OK           string
	InLove       string
	Blowkiss     string
	Waddle       string
	Tremble      string
	Aaagh        string
	Twirl        string
	Kotow        string
	Dramatic     string
	JumpRope     string
	Surrender    string
	Hooray       string
	Meditate     string
	Smooch       string
	TaiChiL      string
	TaiChiR      string
	Hey          string
	Facepalm     string
	Smirk        string
	Smart        string
	Moue         string
	Yeah         string
	Tea          string
	Packet       string
	Candle       string
	Blessing     string
	Chick        string
	Onlooker     string
	GoForIt      string
	Sweats       string
	OMG          string
	Emm          string
	Respect      string
	Doge         string
	NoProb       string
	MyBad        string
	KeepFighting string
	Wow          string
	Rich         string
	Broken       string
	Hurt         string
	Sigh         string
	LetMeSee     string
	Awesome      string
	Boring       string
}{
	Smile:        "[微笑]",
	Grimace:      "[撇嘴]",
	Drool:        "[色]",
	Scowl:        "[发呆]",
	CoolGuy:      "[得意]",
	Sob:          "[流泪]",
	Shy:          "[害羞]",
	Silent:       "[闭嘴]",
	Sleep:        "[睡]",
	Cry:          "[大哭]",
	Awkward:      "[尴尬]",
	Angry:        "[发怒]",
	Tongue:       "[调皮]",
	Grin:         "[呲牙]",
	Surprise:     "[惊讶]",
	Frown:        "[难过]",
	Ruthless:     "[酷]",
	Blush:        "[冷汗]",
	Scream:       "[抓狂]",
	Puke:         "[吐]",
	Chuckle:      "[偷笑]",
	Joyful:       "[愉快]",
	Slight:       "[白眼]",
	Smug:         "[傲慢]",
	Hungry:       "[饥饿]",
	Drowsy:       "[困]",
	Panic:        "[惊恐]",
	Sweat:        "[流汗]",
	Laugh:        "[憨笑]",
	Commando:     "[悠闲]",
	Determined:   "[奋斗]",
	Scold:        "[咒骂]",
	Shocked:      "[疑问]",
	Shhh:         "[嘘]",
	Dizzy:        "[晕]",
	Tormented:    "[疯了]",
	Toasted:      "[衰]",
	Skull:        "[骷髅]",
	Hammer:       "[敲打]",
	Wave:         "[再见]",
	Speechless:   "[擦汗]",
	NosePick:     "[抠鼻]",
	Clap:         "[鼓掌]",
	Shame:        "[糗大了]",
	Trick:        "[坏笑]",
	BahL:         "[左哼哼]",
	BahR:         "[右哼哼]",
	Yawn:         "[哈欠]",
	PoohPooh:     "[鄙视]",
	Shrunken:     "[委屈]",
	TearingUp:    "[快哭了]",
	Sly:          "[阴险]",
	Kiss:         "[亲亲]",
	Wrath:        "[吓]",
	Whimper:      "[可怜]",
	Cleaver:      "[菜刀]",
	Watermelon:   "[西瓜]",
	Beer:         "[啤酒]",
	Basketball:   "[篮球]",
	PingPong:     "[乒乓]",
	Coffee:       "[咖啡]",
	Rice:         "[饭]",
	Pig:          "[猪头]",
	Rose:         "[玫瑰]",
	Wilt:         "[凋谢]",
	Lips:         "[嘴唇]",
	Heart:        "[爱心]",
	BrokenHeart:  "[心碎]",
	Cake:         "[蛋糕]",
	Lightning:    "[闪电]",
	Bomb:         "[炸弹]",
	Dagger:       "[刀]",
	Soccer:       "[足球]",
	Ladybug:      "[瓢虫]",
	Poop:         "[便便]",
	Moon:         "[月亮]",
	Sun:          "[太阳]",
	Gift:         "[礼物]",
	Hug:          "[拥抱]",
	ThumbsUp:     "[强]",
	ThumbsDown:   "[弱]",
	Shake:        "[握手]",
	Peace:        "[胜利]",
	Fight:        "[抱拳]",
	Beckon:       "[勾引]",
	Fist:         "[拳头]",
	Pinky:        "[差劲]",
	RockOn:       "[爱你]",
	Nuhuh:        "[NO]",
	OK:           "[OK]",
	InLove:       "[爱情]",
	Blowkiss:     "[飞吻]",
	Waddle:       "[跳跳]",
	Tremble:      "[发抖]",
	Aaagh:        "[怄火]",
	Twirl:        "[转圈]",
	Kotow:        "[磕头]",
	Dramatic:     "[回头]",
	JumpRope:     "[跳绳]",
	Surrender:    "[投降]",
	Hooray:       "[激动]",
	Meditate:     "[乱舞]",
	Smooch:       "[献吻]",
	TaiChiL:      "[左太极]",
	TaiChiR:      "[右太极]",
	Hey:          "[嘿哈]",
	Facepalm:     "[捂脸]",
	Smirk:        "[奸笑]",
	Smart:        "[机智]",
	Moue:         "[皱眉]",
	Yeah:         "[耶]",
	Tea:          "[茶]",
	Packet:       "[红包]",
	Candle:       "[蜡烛]",
	Blessing:     "[福]",
	Chick:        "[鸡]",
	Onlooker:     "[吃瓜]",
	GoForIt:      "[加油]",
	Sweats:       "[汗]",
	OMG:          "[天啊]",
	Emm:          "[Emm]",
	Respect:      "[社会社会]",
	Doge:         "[旺柴]",
	NoProb:       "[好的]",
	MyBad:        "[打脸]",
	KeepFighting: "[加油加油]",
	Wow:          "[哇]",
	Rich:         "[發]",
	Broken:       "[裂开]",
	Hurt:         "[苦涩]",
	Sigh:         "[叹气]",
	LetMeSee:     "[让我看看]",
	Awesome:      "[666]",
	Boring:       "[翻白眼]",
}

func FormatEmoji(text string) string {
	result := emojiRegexp.FindAllStringSubmatch(text, -1)

	for _, item := range result {
		if len(item) != 2 {
			continue
		}
		value := item[0]
		emojiCodeStr := item[1]
		emojiCode, err := strconv.ParseInt(emojiCodeStr, 16, 64)
		if err != nil {
			continue
		}
		text = strings.Replace(text, value, fmt.Sprintf("%c", emojiCode), -1)
	}

	return text
}
