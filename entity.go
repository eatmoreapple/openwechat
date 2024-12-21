package openwechat

import (
	"errors"
	"fmt"
	"net/url"
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

func (l LoginInfo) Err() error {
	if l.Ok() {
		return nil
	}
	return errors.New(l.Message)
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
	SyncKey             *SyncKey
	User                *User
	MPSubscribeMsgList  []*MPSubscribeMsg
	ContactList         Members
}

// MPSubscribeMsg 公众号的订阅信息
type MPSubscribeMsg struct {
	MPArticleCount int
	Time           int64
	UserName       string
	NickName       string
	MPArticleList  []*MPArticle
}

type MPArticle struct {
	Title  string
	Cover  string
	Digest string
	Url    string
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

type WebWxSyncResponse struct {
	AddMsgCount            int
	ContinueFlag           int
	DelContactCount        int
	ModChatRoomMemberCount int
	ModContactCount        int
	Skey                   string
	SyncCheckKey           SyncKey
	SyncKey                *SyncKey
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

// CheckLoginResponse 检查登录状态的响应body
type CheckLoginResponse []byte

// RedirectURL 重定向的URL
func (c CheckLoginResponse) RedirectURL() (*url.URL, error) {
	code, err := c.Code()
	if err != nil {
		return nil, err
	}
	if code != LoginCodeSuccess {
		return nil, fmt.Errorf("expect status code %s, but got %s", LoginCodeSuccess, code)
	}
	results := redirectUriRegexp.FindSubmatch(c)
	if len(results) != 2 {
		return nil, errors.New("redirect url does not match")
	}
	return url.Parse(string(results[1]))
}

// Code 获取当前的登录检查状态的代码
func (c CheckLoginResponse) Code() (LoginCode, error) {
	results := statusCodeRegexp.FindSubmatch(c)
	if len(results) != 2 {
		return "", errors.New("error status code match")
	}
	code := string(results[1])
	return LoginCode(code), nil
}

// Avatar 获取扫码后的用户头像, base64编码
func (c CheckLoginResponse) Avatar() (string, error) {
	code, err := c.Code()
	if err != nil {
		return "", err
	}
	if code != LoginCodeScanned {
		return "", nil
	}
	results := avatarRegexp.FindSubmatch(c)
	if len(results) != 2 {
		return "", errors.New("avatar does not match")
	}
	return string(results[1]), nil
}

type MessageResponse struct {
	BaseResponse BaseResponse
	LocalID      string
	MsgID        string
}

type UploadResponse struct {
	BaseResponse BaseResponse `json:"BaseResponse"`
	MediaId      string       `json:"MediaId"`
	Signature    string       `json:"Signature"`
}

type PushLoginResponse struct {
	Ret  string `json:"ret"`
	Msg  string `json:"msg"`
	UUID string `json:"uuid"`
}

func (p PushLoginResponse) Ok() bool {
	return p.Ret == "0" && p.UUID != ""
}

func (p PushLoginResponse) Err() error {
	if p.Ok() {
		return nil
	}
	return errors.New(p.Msg)
}
