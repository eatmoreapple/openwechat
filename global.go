package openwechat

import (
	"errors"
	"regexp"
)

var (
	uuidRegexp        = regexp.MustCompile(`uuid = "(.*?)";`)
	statusCodeRegexp  = regexp.MustCompile(`window.code=(\d+);`)
	syncCheckRegexp   = regexp.MustCompile(`window.synccheck=\{retcode:"(\d+)",selector:"(\d+)"\}`)
	redirectUriRegexp = regexp.MustCompile(`window.redirect_uri="(.*?)"`)
)

const (
	appId           = "wx782c26e4c19acffb"
	appMessageAppId = "wxeb7ec651dd0aefa9"

	jsonContentType       = "application/json; charset=utf-8"
	uosPatchClientVersion = "2.0.0"
	uosPatchExtspam       = "Gp8ICJkIEpkICggwMDAwMDAwMRAGGoAI1GiJSIpeO1RZTq9QBKsRbPJdi84ropi16EYI10WB6g74sGmRwSNXjPQnYU" +
		"KYotKkvLGpshucCaeWZMOylnc6o2AgDX9grhQQx7fm2DJRTyuNhUlwmEoWhjoG3F0ySAWUsEbH3bJMsEBwoB//0qmFJob74ffdaslqL+IrSy7L" +
		"J76/G5TkvNC+J0VQkpH1u3iJJs0uUYyLDzdBIQ6Ogd8LDQ3VKnJLm4g/uDLe+G7zzzkOPzCjXL+70naaQ9medzqmh+/SmaQ6uFWLDQLcRln++w" +
		"BwoEibNpG4uOJvqXy+ql50DjlNchSuqLmeadFoo9/mDT0q3G7o/80P15ostktjb7h9bfNc+nZVSnUEJXbCjTeqS5UYuxn+HTS5nZsPVxJA2O5G" +
		"dKCYK4x8lTTKShRstqPfbQpplfllx2fwXcSljuYi3YipPyS3GCAqf5A7aYYwJ7AvGqUiR2SsVQ9Nbp8MGHET1GxhifC692APj6SJxZD3i1drSY" +
		"ZPMMsS9rKAJTGz2FEupohtpf2tgXm6c16nDk/cw+C7K7me5j5PLHv55DFCS84b06AytZPdkFZLj7FHOkcFGJXitHkX5cgww7vuf6F3p0yM/W73" +
		"SoXTx6GX4G6Hg2rYx3O/9VU2Uq8lvURB4qIbD9XQpzmyiFMaytMnqxcZJcoXCtfkTJ6pI7a92JpRUvdSitg967VUDUAQnCXCM/m0snRkR9LtoX" +
		"AO1FUGpwlp1EfIdCZFPKNnXMeqev0j9W9ZrkEs9ZWcUEexSj5z+dKYQBhIICviYUQHVqBTZSNy22PlUIeDeIs11j7q4t8rD8LPvzAKWVqXE+5l" +
		"S1JPZkjg4y5hfX1Dod3t96clFfwsvDP6xBSe1NBcoKbkyGxYK0UvPGtKQEE0Se2zAymYDv41klYE9s+rxp8e94/H8XhrL9oGm8KWb2RmYnAE7r" +
		"y9gd6e8ZuBRIsISlJAE/e8y8xFmP031S6Lnaet6YXPsFpuFsdQs535IjcFd75hh6DNMBYhSfjv456cvhsb99+fRw/KVZLC3yzNSCbLSyo9d9BI" +
		"45Plma6V8akURQA/qsaAzU0VyTIqZJkPDTzhuCl92vD2AD/QOhx6iwRSVPAxcRFZcWjgc2wCKh+uCYkTVbNQpB9B90YlNmI3fWTuUOUjwOzQRx" +
		"JZj11NsimjOJ50qQwTTFj6qQvQ1a/I+MkTx5UO+yNHl718JWcR3AXGmv/aa9rD1eNP8ioTGlOZwPgmr2sor2iBpKTOrB83QgZXP+xRYkb4zVC+" +
		"LoAXEoIa1+zArywlgREer7DLePukkU6wHTkuSaF+ge5Of1bXuU4i938WJHj0t3D8uQxkJvoFi/EYN/7u2P1zGRLV4dHVUsZMGCCtnO6BBigFMAA="
)

// 消息类型
const (
	TextMessage  = 1
	ImageMessage = 3
	AppMessage   = 6
)

// https://res.wx.qq.com/a/wx_fed/webwx/res/static/js/index_c7d281c.js
// varcaser.Caser{
//		From: varcaser.ScreamingSnakeCase, To: varcaser.UpperCamelCaseKeepCaps}
const (
	MsgtypeText              = 1     // 文本消息
	MsgtypeImage             = 3     // 图片消息
	MsgtypeVoice             = 34    // 语音消息
	MsgtypeVerifymsg         = 37    // 认证消息
	MsgtypePossiblefriendMsg = 40    // 好友推荐消息
	MsgtypeSharecard         = 42    // 名片消息
	MsgtypeVideo             = 43    // 视频消息
	MsgtypeEmoticon          = 47    // 表情消息
	MsgtypeLocation          = 48    // 地理位置消息
	MsgtypeApp               = 49    // APP消息
	MsgtypeVoipmsg           = 50    // voip msg	//VOIP消息
	MsgtypeVoipnotify        = 52    // voip 结束消息
	MsgtypeVoipinvite        = 53    // voip 邀请
	MsgtypeMicrovideo        = 62    // 小视频消息
	MsgtypeSys               = 10000 // 系统消息
	MsgtypeRecalled          = 10002 // 消息撤回
)

// 登录状态
const (
	statusSuccess = "200"
	statusScanned = "201"
	statusTimeout = "400"
	statusWait    = "408"
)

// errors
var (
	noSuchUserFoundError = errors.New("no such user found")
	missLocationHeader   = errors.New("301 response missing Location header")
	loginForbiddenError  = errors.New("login forbidden")
)

// ALL 跟search函数搭配
//      friends.Search(openwechat.ALL, )
const ALL = 0

// 性别
const (
	MALE   = 1
	FEMALE = 2
)

const (
	// 分块上传时每次上传的文件的大小
	chunkSize int64 = (1 << 20) / 2 // 0.5m
	// 需要检测的文件大小
	needCheckSize int64 = 25 << 20 // 20m
	// 最大文件上传大小
	maxFileUploadSize int64 = 50 << 20 // 50m
	// 最大图片上传大小
	maxImageUploadSize int64 = 20 << 20 // 20m
)

const TimeFormat = "Mon Jan 02 2006 15:04:05 GMT+0800 (中国标准时间)"

var imageType = map[string]bool{
	"bmp":  true,
	"png":  true,
	"jpeg": true,
	"jpg":  true,
}

var videoType = "mp4"

// ZombieText 检测僵尸好友字符串
// 发送该字符给好友，能正常发送不报错的为正常好友，否则为僵尸好友
const ZombieText = "وُحfخe ̷̴̐nخg ̷̴̐cخh ̷̴̐aخo امارتيخ ̷̴̐خ\n"
