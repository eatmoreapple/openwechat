package openwechat

import (
	"fmt"
	"regexp"
	"testing"
)

func ExampleMessageType_output() {
	for _, wxt := range []MessageType{
		MsgTypeText, MsgTypeImage, MsgTypeVoice, MsgTypeVerify,
		MsgTypePossibleFriend, MsgTypeShareCard, MsgTypeVideo, MsgTypeEmoticon,
		MsgTypeLocation, MsgTypeApp, MsgTypeVoip, MsgTypeVoipNotify,
		MsgTypeVoipInvite, MsgTypeMicroVideo, MsgTypeSys, MsgTypeRecalled} {
		fmt.Printf("收到一条%s(type %d)\n", wxt, wxt)
	}
	fmt.Println("=======")
	for _, wxt := range []MessageType{10000, 6, 51} {
		wxtstr := wxt.String()
		if regexp.MustCompile(`^M`).MatchString(wxtstr) {
			wxtstr = "未知消息"
		}
		fmt.Printf("收到一条%s(type %d): %s\n", wxtstr, wxt, wxt)
	}
	// Output:
	// 收到一条文本消息(type 1)
	// 收到一条图片消息(type 3)
	// 收到一条语音消息(type 34)
	// 收到一条认证消息(type 37)
	// 收到一条好友推荐消息(type 40)
	// 收到一条名片消息(type 42)
	// 收到一条视频消息(type 43)
	// 收到一条表情消息(type 47)
	// 收到一条地理位置消息(type 48)
	// 收到一条APP消息(type 49)
	// 收到一条VOIP消息(type 50)
	// 收到一条VOIP结束消息(type 52)
	// 收到一条VOIP邀请(type 53)
	// 收到一条小视频消息(type 62)
	// 收到一条系统消息(type 10000)
	// 收到一条消息撤回(type 10002)
	// =======
	// 收到一条系统消息(type 10000): 系统消息
	// 收到一条未知消息(type 6): MessageType(6)
	// 收到一条未知消息(type 51): MessageType(51)
}

func TestMessage_FriendAddMessageContent(t *testing.T) {
	m := &Message{
		MsgType:      37,
		FromUserName: "fmessage",
		Content:      "<msg fromusername=\"user1\" encryptusername=\"123123131@stranger\" fromnickname=\"nickname1\" content=\"test11111\"  shortpy=\"MMHDNLXX\" imagestatus=\"3\" scene=\"3\" country=\"CN\" province=\"Zhejiang\" city=\"Hangzhou\" sign=\"Life is a struggle.\" percard=\"1\" sex=\"1\" alias=\"1111\" weibo=\"\" albumflag=\"0\" albumstyle=\"0\" albumbgimgid=\"\" snsflag=\"273\" snsbgimgid=\"http://shmmsns.qpic.cn/mbsFs/0\" snsbgobjectid=\"123123131\" mhash=\"1231231\" mfullhash=\"1231231\" bigheadimgurl=\"http://wx.qlogo.cn/mmhead/ver_1/1231231/0\" smallheadimgurl=\"http://wx.qlogo.cn/mmhead/ver_1/1231231/132\" ticket=\"1231231@stranger\" opcode=\"2\" googlecontact=\"\" qrticket=\"\" chatroomusername=\"123@chatroom\" sourceusername=\"\" sourcenickname=\"\" sharecardusername=\"\" sharecardnickname=\"\" cardversion=\"\" extflag=\"0\"><brandlist count=\"0\" ver=\"759291380\"></brandlist></msg>",
	}
	fm, err := m.FriendAddMessageContent()
	if err != nil {
		t.Error(err)
	}

	if fm.Shortpy != "MMHDNLXX" {
		t.Error("unexpected shorty:", fm.Shortpy)
	}
}
