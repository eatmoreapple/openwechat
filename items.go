package openwechat

import (
	"errors"
	"fmt"
)

/*
一些网络返回信息的封装
*/

// 登录信息
type LoginInfo struct {
	Ret         int    `xml:"ret"`
	WxUin       int    `xml:"wxuin"`
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

// 初始的请求信息
// 几乎所有的请求都要携带该参数
type BaseRequest struct {
	Uin                 int
	Sid, Skey, DeviceID string
}

// 大部分返回对象都携带该信息
type BaseResponse struct {
	Ret    int
	ErrMsg string
}

func (b BaseResponse) Ok() bool {
	return b.Ret == 0
}

func (b BaseResponse) Error() string {
	if err := getResponseErrorWithRetCode(b.Ret); err != nil {
		return err.Error()
	}
	return ""
}

func getResponseErrorWithRetCode(code int) error {
	switch code {
	case 0:
		return nil
	case 1:
		return errors.New("param error")
	case -14:
		return errors.New("ticket error")
	case 1100:
		return errors.New("not login warn")
	case 1101:
		return errors.New("not login check")
	case 1102:
		return errors.New("cookie invalid error")
	case 1203:
		return errors.New("login env error")
	case 1205:
		return errors.New("opt too often")
	default:
		return fmt.Errorf("base response ret code %d", code)
	}
}

type SyncKey struct {
	Count int
	List  []struct{ Key, Val int64 }
}

// 初始化的相应信息
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

// 公众号的订阅信息
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
		return "opt too often"
	default:
		return fmt.Sprintf("sync check response error code %s", s.RetCode)
	}
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
	Seq          int
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
