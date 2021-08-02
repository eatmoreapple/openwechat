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
	AppMessage = 6
)

// MessageType以Go惯用形式定义了PC微信所有的官方消息类型。
// 详见 message_test.go
type MessageType int

// AppMessageType以Go惯用形式定义了PC微信所有的官方App消息类型。
type AppMessageType int

//go:generate stringer -type=MessageType -linecomment=true
//go:generate stringer -type=AppMessageType -linecomment=true

// https://res.wx.qq.com/a/wx_fed/webwx/res/static/js/index_c7d281c.js
// MSGTYPE_TEXT
// MSGTYPE_IMAGE
// MSGTYPE_VOICE
// MSGTYPE_VERIFYMSG
// MSGTYPE_POSSIBLEFRIEND_MSG
// MSGTYPE_SHARECARD
// MSGTYPE_VIDEO
// MSGTYPE_EMOTICON
// MSGTYPE_LOCATION
// MSGTYPE_APP
// MSGTYPE_VOIPMSG
// MSGTYPE_VOIPNOTIFY
// MSGTYPE_VOIPINVITE
// MSGTYPE_MICROVIDEO
// MSGTYPE_SYS
// MSGTYPE_RECALLED

const (
	MsgTypeText           MessageType = 1     // 文本消息
	MsgTypeImage          MessageType = 3     // 图片消息
	MsgTypeVoice          MessageType = 34    // 语音消息
	MsgTypeVerify         MessageType = 37    // 认证消息
	MsgTypePossibleFriend MessageType = 40    // 好友推荐消息
	MsgTypeShareCard      MessageType = 42    // 名片消息
	MsgTypeVideo          MessageType = 43    // 视频消息
	MsgTypeEmoticon       MessageType = 47    // 表情消息
	MsgTypeLocation       MessageType = 48    // 地理位置消息
	MsgTypeApp            MessageType = 49    // APP消息
	MsgTypeVoip           MessageType = 50    // VOIP消息
	MsgTypeVoipNotify     MessageType = 52    // VOIP结束消息
	MsgTypeVoipInvite     MessageType = 53    // VOIP邀请
	MsgTypeMicroVideo     MessageType = 62    // 小视频消息
	MsgTypeSys            MessageType = 10000 // 系统消息
	MsgTypeRecalled       MessageType = 10002 // 消息撤回
)

const (
	AppMsgTypeText                  AppMessageType = 1      // 文本消息
	AppMsgTypeImg                   AppMessageType = 2      // 图片消息
	AppMsgTypeAudio                 AppMessageType = 3      // 语音消息
	AppMsgTypeVideo                 AppMessageType = 4      // 视频消息
	AppMsgTypeUrl                   AppMessageType = 5      // 文章消息
	AppMsgTypeAttach                AppMessageType = 6      // 附件消息
	AppMsgTypeOpen                  AppMessageType = 7      // Open
	AppMsgTypeEmoji                 AppMessageType = 8      // 表情消息
	AppMsgTypeVoiceRemind           AppMessageType = 9      // VoiceRemind
	AppMsgTypeScanGood              AppMessageType = 10     // ScanGood
	AppMsgTypeGood                  AppMessageType = 13     // Good
	AppMsgTypeEmotion               AppMessageType = 15     // Emotion
	AppMsgTypeCardTicket            AppMessageType = 16     // 名片消息
	AppMsgTypeRealtimeShareLocation AppMessageType = 17     // 地理位置消息
	AppMsgTypeTransfers             AppMessageType = 2000   // 转账消息
	AppMsgTypeRedEnvelopes          AppMessageType = 2001   // 红包消息
	AppMsgTypeReaderType            AppMessageType = 100001 //自定义的消息

)

// 登录状态
const (
	StatusSuccess = "200"
	StatusScanned = "201"
	StatusTimeout = "400"
	StatusWait    = "408"
)

// errors
var (
	ErrNoSuchUserFoundError = errors.New("no such user found")
	ErrMissLocationHeader   = errors.New("301 response missing Location header")
	ErrLoginForbiddenError  = errors.New("login forbidden")
	ErrLoginTimeout         = errors.New("login timeout")
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
