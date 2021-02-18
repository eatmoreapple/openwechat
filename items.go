package openwechat

import "fmt"

type LoginInfo struct {
	Ret         int    `xml:"ret"`
	Message     string `xml:"message"`
	SKey        string `xml:"skey"`
	WxSid       string `xml:"wxsid"`
	WxUin       int    `xml:"wxuin"`
	PassTicket  string `xml:"pass_ticket"`
	IsGrayScale int    `xml:"isgrayscale"`
}

type BaseRequest struct {
	Uin                 int
	Sid, Skey, DeviceID string
}

type BaseResponse struct {
	ErrMsg string
	Ret    int
}

func (b BaseResponse) Ok() bool {
	return b.Ret == 0
}

// 实现error接口
func (b BaseResponse) Error() string {
	switch b.Ret {
	case 0:
		return ""
	case 1:
		return "param error"
	case -14:
		return "ticker error"
	case 1100:
		return "not login warn"
	case 1101:
		return "not login check"
	case 1102:
		return "cookie invalid error"
	case 1203:
		return "login env error"
	case 1205:
		return "op too often"
	default:
		if b.ErrMsg != "" {
			return b.ErrMsg
		}
		return fmt.Sprintf("base response error code %d", b.Ret)
	}
}

type SyncKey struct {
	Count int
	List  []struct{ Key, Val int64 }
}

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
	list := make(UserDetailItemList, members.Count()-1)
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
