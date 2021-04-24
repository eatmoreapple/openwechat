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
	appId = "wx782c26e4c19acffb"

	baseUrl                 = "https://wx.qq.com"
	jsLoginUrl              = "https://login.wx.qq.com/jslogin"
	webWxNewLoginPage       = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?mod=desktop"
	qrcodeUrl               = "https://login.weixin.qq.com/qrcode/"
	loginUrl                = "https://login.wx.qq.com/cgi-bin/mmwebwx-bin/login"
	webWxInitUrl            = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxinit"
	webWxStatusNotifyUrl    = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxstatusnotify"
	webWxSyncUrl            = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsync"
	webWxSendMsgUrl         = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsg"
	webWxGetContactUrl      = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetcontact"
	webWxSendMsgImgUrl      = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsgimg"
	webWxSendAppMsgUrl      = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsendappmsg"
	webWxBatchGetContactUrl = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxbatchgetcontact"
	webWxOplogUrl           = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxoplog"
	webWxVerifyUserUrl      = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxverifyuser"
	syncCheckUrl            = "https://webpush.wx.qq.com/cgi-bin/mmwebwx-bin/synccheck"
	webWxUpLoadMediaUrl     = "https://file.wx.qq.com/cgi-bin/mmwebwx-bin/webwxuploadmedia"
	webWxGetMsgImgUrl       = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetmsgimg"
	webWxGetVoiceUrl        = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetvoice"
	webWxGetVideoUrl        = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetvideo"
	webWxLogoutUrl          = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxlogout"
	webWxGetMediaUrl        = "https://file.wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetmedia"

	jsonContentType       = "application/json; charset=utf-8"
	uosPatchClientVersion = "2.0.0"
	uosPatchExtspam       = "Gp8ICJkIEpkICggwMDAwMDAwMRAGGoAI1GiJSIpeO1RZTq9QBKsRbPJdi84ropi16EYI10WB6g74sGmRwSNXjPQnYUKYotKkvLGpshucCaeWZMOylnc6o2AgDX9grhQQx7fm2DJRTyuNhUlwmEoWhjoG3F0ySAWUsEbH3bJMsEBwoB//0qmFJob74ffdaslqL+IrSy7LJ76/G5TkvNC+J0VQkpH1u3iJJs0uUYyLDzdBIQ6Ogd8LDQ3VKnJLm4g/uDLe+G7zzzkOPzCjXL+70naaQ9medzqmh+/SmaQ6uFWLDQLcRln++wBwoEibNpG4uOJvqXy+ql50DjlNchSuqLmeadFoo9/mDT0q3G7o/80P15ostktjb7h9bfNc+nZVSnUEJXbCjTeqS5UYuxn+HTS5nZsPVxJA2O5GdKCYK4x8lTTKShRstqPfbQpplfllx2fwXcSljuYi3YipPyS3GCAqf5A7aYYwJ7AvGqUiR2SsVQ9Nbp8MGHET1GxhifC692APj6SJxZD3i1drSYZPMMsS9rKAJTGz2FEupohtpf2tgXm6c16nDk/cw+C7K7me5j5PLHv55DFCS84b06AytZPdkFZLj7FHOkcFGJXitHkX5cgww7vuf6F3p0yM/W73SoXTx6GX4G6Hg2rYx3O/9VU2Uq8lvURB4qIbD9XQpzmyiFMaytMnqxcZJcoXCtfkTJ6pI7a92JpRUvdSitg967VUDUAQnCXCM/m0snRkR9LtoXAO1FUGpwlp1EfIdCZFPKNnXMeqev0j9W9ZrkEs9ZWcUEexSj5z+dKYQBhIICviYUQHVqBTZSNy22PlUIeDeIs11j7q4t8rD8LPvzAKWVqXE+5lS1JPZkjg4y5hfX1Dod3t96clFfwsvDP6xBSe1NBcoKbkyGxYK0UvPGtKQEE0Se2zAymYDv41klYE9s+rxp8e94/H8XhrL9oGm8KWb2RmYnAE7ry9gd6e8ZuBRIsISlJAE/e8y8xFmP031S6Lnaet6YXPsFpuFsdQs535IjcFd75hh6DNMBYhSfjv456cvhsb99+fRw/KVZLC3yzNSCbLSyo9d9BI45Plma6V8akURQA/qsaAzU0VyTIqZJkPDTzhuCl92vD2AD/QOhx6iwRSVPAxcRFZcWjgc2wCKh+uCYkTVbNQpB9B90YlNmI3fWTuUOUjwOzQRxJZj11NsimjOJ50qQwTTFj6qQvQ1a/I+MkTx5UO+yNHl718JWcR3AXGmv/aa9rD1eNP8ioTGlOZwPgmr2sor2iBpKTOrB83QgZXP+xRYkb4zVC+LoAXEoIa1+zArywlgREer7DLePukkU6wHTkuSaF+ge5Of1bXuU4i938WJHj0t3D8uQxkJvoFi/EYN/7u2P1zGRLV4dHVUsZMGCCtnO6BBigFMAA="
)

// 消息类型
const (
	TextMessage  = 1
	ImageMessage = 3
	AppMessage   = 6
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
)

const ALL = 0

// sex
const (
	MALE   = 1
	FEMALE = 2
)

type mode string

const (
	normal mode = "normal"
	desk   mode = "desk"
)
