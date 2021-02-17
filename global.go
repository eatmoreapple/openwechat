package openwechat

import "regexp"

var (
	uuidRegexp        = regexp.MustCompile(`uuid = "(.*?)";`)
	statusCodeRegexp  = regexp.MustCompile(`window.code=(\d+);`)
	syncCheckRegexp   = regexp.MustCompile(`window.synccheck=\{retcode:"(\d+)",selector:"(\d+)"\}`)
	redirectUriRegexp = regexp.MustCompile(`window.redirect_uri="(.*?)"`)
)

const (
	appId = "wx782c26e4c19acffb"

	baseUrl                 = "https://wx2.qq.com"
	jsLoginUrl              = "https://login.wx.qq.com/jslogin"
	webWxNewLoginPage       = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage"
	qrcodeUrl               = "https://login.weixin.qq.com/qrcode/"
	loginUrl                = "https://login.wx.qq.com/cgi-bin/mmwebwx-bin/login"
	webWxInitUrl            = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxinit"
	webWxStatusNotifyUrl    = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxstatusnotify"
	webWxSyncUrl            = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxsync"
	webWxSendMsgUrl         = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsg"
	webWxGetContactUrl      = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxgetcontact"
	webWxSendMsgImgUrl      = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsgimg"
	webWxSendAppMsgUrl      = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxsendappmsg"
	webWxBatchGetContactUrl = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxbatchgetcontact"
	webWxOplogUrl           = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxoplog"
	webWxVerifyUserUrl      = "https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxverifyuser"
	syncCheckUrl            = "https://webpush.wx2.qq.com/cgi-bin/mmwebwx-bin/synccheck"
	webWxUpLoadMediaUrl     = "https://file.wx2.qq.com/cgi-bin/mmwebwx-bin/webwxuploadmedia"

	jsonContentType = "application/json; charset=utf-8"
)

const (
	TextMessage  = 1
	ImageMessage = 3
	AppMessage   = 6
)

const (
	statusSuccess = "200"
	statusScanned = "201"
	statusTimeout = "400"
	statusWait    = "408"
)
