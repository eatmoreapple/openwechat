package openwechat

import (
	"fmt"
)

func ExampleWxMsgType_output() {
	for _, wxt := range []int{
		MsgtypeText, MsgtypeImage, MsgtypeVoice, MsgtypeVerifymsg,
		MsgtypePossiblefriendMsg, MsgtypeSharecard, MsgtypeVideo, MsgtypeEmoticon,
		MsgtypeLocation, MsgtypeApp, MsgtypeVoipmsg, MsgtypeVoipnotify,
		MsgtypeVoipinvite, MsgtypeMicrovideo, MsgtypeSys, MsgtypeRecalled} {
		fmt.Printf("收到一条%s(type %d)\n", WxMsgType.String(wxt), wxt)
	}
	fmt.Println("=======")
	for _, wxt := range []int{10000, 6, 51} {
		wxtstr, ok := WxMsgType.Exist(wxt)
		if !ok {
			wxtstr = "未知消息"
		}
		fmt.Printf("收到一条%s(type %d)\n", wxtstr, wxt)
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
	// 收到一条系统消息(type 10000)
	// 收到一条未知消息(type 6)
	// 收到一条未知消息(type 51)
}
