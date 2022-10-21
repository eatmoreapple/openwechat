package openwechat

import (
	"strconv"
)

/*
一些网络返回信息的封装
*/

// LoginInfo 登录信息
type LoginInfo struct {
	Ret         int    `xml:"ret"`
	WxUin       int64  `xml:"wxuin"`
	IsGrayScale int    `xml:"isgrayscale"`
	Message     string `xml:"message"`
	SKey        string `xml:"skey"`
	WxSid       string `xml:"wxsid"`
	PassTicket  string `xml:"pass_ticket"`
}

func (l LoginInfo) Ok() bool {
	return l.Ret == 0
}

func (l LoginInfo) Error() string {
	return l.Message
}

// BaseRequest 初始的请求信息
// 几乎所有的请求都要携带该参数
type BaseRequest struct {
	Uin                 int64
	Sid, Skey, DeviceID string
}

type SyncKey struct {
	Count int
	List  []struct{ Key, Val int64 }
}

// WebInitResponse 初始化的相应信息
type WebInitResponse struct {
	Count               int
	ClientVersion       int
	GrayScale           int
	InviteStartCount    int
	MPSubscribeMsgCount int
	ClickReportInterval int
	SystemTime          int64
	ChatSet             string
	SKey                string
	BaseResponse        BaseResponse
	SyncKey             SyncKey
	User                User
	MPSubscribeMsgList  []MPSubscribeMsg
	ContactList         []User
}

// MPSubscribeMsg 公众号的订阅信息
type MPSubscribeMsg struct {
	MPArticleCount int
	Time           int64
	UserName       string
	NickName       string
	MPArticleList  []struct {
		Title  string
		Cover  string
		Digest string
		Url    string
	}
}

type UserDetailItem struct {
	UserName        string
	EncryChatRoomId string
}

type UserDetailItemList []UserDetailItem

func NewUserDetailItemList(members Members) UserDetailItemList {
	var list = make(UserDetailItemList, len(members))
	for index, member := range members {
		item := UserDetailItem{UserName: member.UserName, EncryChatRoomId: member.EncryChatRoomId}
		list[index] = item
	}
	return list
}

type SyncCheckResponse struct {
	RetCode  string
	Selector string
}

func (s SyncCheckResponse) Success() bool {
	return s.RetCode == "0"
}

func (s SyncCheckResponse) NorMal() bool {
	return s.Success() && s.Selector == "0"
}

// 实现error接口
func (s SyncCheckResponse) Error() string {
	i, err := strconv.ParseInt(s.RetCode, 16, 10)
	if err != nil {
		return ""
	}
	return Ret(i).String()
}

type WebWxSyncResponse struct {
	AddMsgCount            int
	ContinueFlag           int
	DelContactCount        int
	ModChatRoomMemberCount int
	ModContactCount        int
	Skey                   string
	SyncCheckKey           SyncKey
	SyncKey                SyncKey
	BaseResponse           BaseResponse
	ModChatRoomMemberList  Members
	AddMsgList             []*Message
}

type WebWxContactResponse struct {
	MemberCount  int
	Seq          int64
	BaseResponse BaseResponse
	MemberList   []*User
}

type WebWxBatchContactResponse struct {
	Count        int
	BaseResponse BaseResponse
	ContactList  []*User
}

type CheckLoginResponse struct {
	Code string
	Raw  []byte
}

type MessageResponse struct {
	BaseResponse BaseResponse
	LocalID      string
	MsgID        string
}

type UploadResponse struct {
	BaseResponse BaseResponse
	MediaId      string
}

type PushLoginResponse struct {
	Ret  string `json:"ret"`
	Msg  string `json:"msg"`
	UUID string `json:"uuid"`
}

func (p PushLoginResponse) Ok() bool {
	return p.Ret == "0" && p.UUID != ""
}
