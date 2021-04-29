package openwechat

// url信息存储
type UrlManager struct {
	baseUrl                 string
	webWxNewLoginPageUrl    string
	webWxInitUrl            string
	webWxStatusNotifyUrl    string
	webWxSyncUrl            string
	webWxSendMsgUrl         string
	webWxGetContactUrl      string
	webWxSendMsgImgUrl      string
	webWxSendAppMsgUrl      string
	webWxBatchGetContactUrl string
	webWxOplogUrl           string
	webWxVerifyUserUrl      string
	syncCheckUrl            string
	webWxUpLoadMediaUrl     string
	webWxGetMsgImgUrl       string
	webWxGetVoiceUrl        string
	webWxGetVideoUrl        string
	webWxLogoutUrl          string
	webWxGetMediaUrl        string
	webWxUpdateChatRoomUrl  string
	webWxRevokeMsg          string
	webWxCheckUploadUrl     string
}

var (
	// uos版
	desktop = UrlManager{
		baseUrl:                 baseDesktopUrl,
		webWxNewLoginPageUrl:    webWxNewLoginPageDesktopUrl,
		webWxInitUrl:            webWxInitDesktopUrl,
		webWxStatusNotifyUrl:    webWxStatusNotifyDesktopUrl,
		webWxSyncUrl:            webWxSyncDesktopUrl,
		webWxSendMsgUrl:         webWxSendMsgDesktopUrl,
		webWxGetContactUrl:      webWxGetContactDesktopUrl,
		webWxSendMsgImgUrl:      webWxSendMsgImgDesktopUrl,
		webWxSendAppMsgUrl:      webWxSendAppMsgDesktopUrl,
		webWxBatchGetContactUrl: webWxBatchGetContactDesktopUrl,
		webWxOplogUrl:           webWxOplogDesktopUrl,
		webWxVerifyUserUrl:      webWxVerifyUserDesktopUrl,
		syncCheckUrl:            syncCheckDesktopUrl,
		webWxUpLoadMediaUrl:     webWxUpLoadMediaDesktopUrl,
		webWxGetMsgImgUrl:       webWxGetMsgImgDesktopUrl,
		webWxGetVoiceUrl:        webWxGetVoiceDesktopUrl,
		webWxGetVideoUrl:        webWxGetVideoDesktopUrl,
		webWxLogoutUrl:          webWxLogoutDesktopUrl,
		webWxGetMediaUrl:        webWxGetMediaDesktopUrl,
		webWxUpdateChatRoomUrl:  webWxUpdateChatRoomDesktopUrl,
		webWxRevokeMsg:          webWxRevokeMsgDesktopUrl,
		webWxCheckUploadUrl:     webWxCheckUploadDesktopUrl,
	}

	// 网页版
	normal = UrlManager{
		baseUrl:                 baseNormalUrl,
		webWxNewLoginPageUrl:    webWxNewLoginPageNormalUrl,
		webWxInitUrl:            webWxInitNormalUrl,
		webWxStatusNotifyUrl:    webWxStatusNotifyNormalUrl,
		webWxSyncUrl:            webWxSyncNormalUrl,
		webWxSendMsgUrl:         webWxSendMsgNormalUrl,
		webWxGetContactUrl:      webWxGetContactNormalUrl,
		webWxSendMsgImgUrl:      webWxSendMsgImgNormalUrl,
		webWxSendAppMsgUrl:      webWxSendAppMsgNormalUrl,
		webWxBatchGetContactUrl: webWxBatchGetContactNormalUrl,
		webWxOplogUrl:           webWxOplogNormalUrl,
		webWxVerifyUserUrl:      webWxVerifyUserNormalUrl,
		syncCheckUrl:            syncCheckNormalUrl,
		webWxUpLoadMediaUrl:     webWxUpLoadMediaNormalUrl,
		webWxGetMsgImgUrl:       webWxGetMsgImgNormalUrl,
		webWxGetVoiceUrl:        webWxGetVoiceNormalUrl,
		webWxGetVideoUrl:        webWxGetVideoNormalUrl,
		webWxLogoutUrl:          webWxLogoutNormalUrl,
		webWxGetMediaUrl:        webWxGetMediaNormalUrl,
		webWxUpdateChatRoomUrl:  webWxUpdateChatRoomNormalUrl,
		webWxRevokeMsg:          webWxRevokeMsgNormalUrl,
		webWxCheckUploadUrl:     webWxCheckUploadNormalUrl,
	}
)

// mode 类型限制
type mode string

// 向外暴露2种模式
const (
	Normal  mode = "normal"
	Desktop mode = "desktop" // 突破网页版登录限制
)

// 通过mode获取完善的UrlManager,
// mode有且仅有两种模式: Normal && Desktop
func GetUrlManagerByMode(m mode) UrlManager {
	switch m {
	case Desktop:
		return desktop
	case Normal:
		return normal
	default:
		panic("unsupport mode got")
	}
}
