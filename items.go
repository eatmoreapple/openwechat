package openwechat

import "fmt"

/*
一些网络返回信息的封装
*/

// 登录信息
type LoginInfo struct {
	Ret         int    `xml:"ret"`
	Message     string `xml:"message"`
	SKey        string `xml:"skey"`
	WxSid       string `xml:"wxsid"`
	WxUin       int    `xml:"wxuin"`
	PassTicket  string `xml:"pass_ticket"`
	IsGrayScale int    `xml:"isgrayscale"`
}

// 初始的请求信息
// 几乎所有的请求都要携带该参数
type BaseRequest struct {
	Uin                 int
	Sid, Skey, DeviceID string
}

// 大部分返回对象都携带该信息
type BaseResponse struct {
	ErrMsg string
	Ret    int
}

func (b BaseResponse) Ok() bool {
	return b.Ret == 0
}

func (b BaseResponse) Error() string {
	switch b.Ret {
	case 0:
		return ""
	case 1:
		return "param error"
	case -14:
		return "ticket error"
	case 1100:
		return "not login warn"
	case 1101:
		return "not login check"
	case 1102:
		return "cookie invalid error"
	case 1203:
		return "login env error"
	case 1205:
		return "opt too often"
	default:
		return fmt.Sprintf("base response ret code %d", b.Ret)
	}
}

type SyncKey struct {
	Count int
	List  []struct{ Key, Val int64 }
}

// 初始化的相应信息
type WebInitResponse struct {
	BaseResponse        BaseResponse
	Count               int
	ChatSet             string
	SKey                string
	SyncKey             SyncKey
	User                User
	ClientVersion       int
	SystemTime          int64
	GrayScale           int
	InviteStartCount    int
	MPSubscribeMsgCount int
	MPSubscribeMsgList  []MPSubscribeMsg
	ClickReportInterval int
	ContactList         []User
}

// 公众号的订阅信息
type MPSubscribeMsg struct {
	UserName       string
	Time           int64
	NickName       string
	MPArticleCount int
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
	list := make(UserDetailItemList, 0)
	for _, member := range members {
		item := UserDetailItem{UserName: member.UserName, EncryChatRoomId: member.EncryChatRoomId}
		list = append(list, item)
	}
	return list
}

type SyncCheckResponse struct {
	RetCode  string
	Selector string
}

func (s *SyncCheckResponse) Success() bool {
	return s.RetCode == "0"
}

func (s *SyncCheckResponse) NorMal() bool {
	return s.Success() && s.Selector == "0"
}

// 实现error接口
func (s *SyncCheckResponse) Error() string {
	switch s.RetCode {
	case "0":
		return ""
	case "1":
		return "param error"
	case "-14":
		return "ticker error"
	case "1100":
		return "not login warn"
	case "1101":
		return "not login check"
	case "1102":
		return "cookie invalid error"
	case "1203":
		return "login env error"
	case "1205":
		return "op too often"
	default:
		return fmt.Sprintf("sync check response error code %s", s.RetCode)
	}
}

type WebWxSyncResponse struct {
	AddMsgCount            int
	AddMsgList             []*Message
	BaseResponse           BaseResponse
	ContinueFlag           int
	DelContactCount        int
	ModChatRoomMemberCount int
	ModChatRoomMemberList  Members
	ModContactCount        int
	Skey                   string
	SyncCheckKey           SyncKey
	SyncKey                SyncKey
}

type WebWxContactResponse struct {
	BaseResponse BaseResponse
	MemberCount  int
	MemberList   []*User
	Seq          int
}

type WebWxBatchContactResponse struct {
	BaseResponse BaseResponse
	ContactList  []*User
	Count        int
}

type CheckLoginResponse struct {
	Code string
	Raw  []byte
}
