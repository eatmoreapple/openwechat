package openwechat

// Session 会话信息，包含登录信息、请求信息、响应信息
type Session struct {
	LoginInfo *LoginInfo
	Request   *BaseRequest
	Response  *WebInitResponse
}
