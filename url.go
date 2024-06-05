package openwechat

//// mode 类型限制
//type mode string
//
//// 向外暴露2种模式
//const (
//	Normal  mode = "normal"
//	Desktop mode = "desktop" // 突破网页版登录限制
//)

const (
	webwxinit            = "/cgi-bin/mmwebwx-bin/webwxinit"
	webwxstatusnotify    = "/cgi-bin/mmwebwx-bin/webwxstatusnotify"
	webwxsync            = "/cgi-bin/mmwebwx-bin/webwxsync"
	webwxsendmsg         = "/cgi-bin/mmwebwx-bin/webwxsendmsg"
	webwxsendemoticon    = "/cgi-bin/mmwebwx-bin/webwxsendemoticon"
	webwxgetcontact      = "/cgi-bin/mmwebwx-bin/webwxgetcontact"
	webwxsendmsgimg      = "/cgi-bin/mmwebwx-bin/webwxsendmsgimg"
	webwxsendappmsg      = "/cgi-bin/mmwebwx-bin/webwxsendappmsg"
	webwxsendvideomsg    = "/cgi-bin/mmwebwx-bin/webwxsendvideomsg"
	webwxbatchgetcontact = "/cgi-bin/mmwebwx-bin/webwxbatchgetcontact"
	webwxoplog           = "/cgi-bin/mmwebwx-bin/webwxoplog"
	webwxverifyuser      = "/cgi-bin/mmwebwx-bin/webwxverifyuser"
	synccheck            = "/cgi-bin/mmwebwx-bin/synccheck"
	webwxuploadmedia     = "/cgi-bin/mmwebwx-bin/webwxuploadmedia"
	webwxgetmsgimg       = "/cgi-bin/mmwebwx-bin/webwxgetmsgimg"
	webwxgetvoice        = "/cgi-bin/mmwebwx-bin/webwxgetvoice"
	webwxgetvideo        = "/cgi-bin/mmwebwx-bin/webwxgetvideo"
	webwxlogout          = "/cgi-bin/mmwebwx-bin/webwxlogout"
	webwxgetmedia        = "/cgi-bin/mmwebwx-bin/webwxgetmedia"
	webwxupdatechatroom  = "/cgi-bin/mmwebwx-bin/webwxupdatechatroom"
	webwxrevokemsg       = "/cgi-bin/mmwebwx-bin/webwxrevokemsg"
	webwxcheckupload     = "/cgi-bin/mmwebwx-bin/webwxcheckupload"
	webwxpushloginurl    = "/cgi-bin/mmwebwx-bin/webwxpushloginurl"
	webwxgeticon         = "/cgi-bin/mmwebwx-bin/webwxgeticon"
	webwxcreatechatroom  = "/cgi-bin/mmwebwx-bin/webwxcreatechatroom"

	webwxnewloginpage = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage"
	jslogin           = "https://login.wx.qq.com/jslogin"
	login             = "https://login.wx.qq.com/cgi-bin/mmwebwx-bin/login"
	qrcode            = "https://login.weixin.qq.com/qrcode/"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "https://" + string(w)
}

func (w WechatDomain) FileHost() string {
	return "https://file." + string(w)
}

func (w WechatDomain) SyncHost() string {
	return "https://webpush." + string(w)
}
