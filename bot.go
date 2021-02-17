package openwechat

import (
	"errors"
	"fmt"
)

type Bot struct {
	Caller               *Caller
	self                 *Self
	storage              WechatStorage
	ScanCallBack         func(body []byte)
	LoginCallBack        func(body []byte)
	UUIDCallback         func(uuid string)
	messageHandlerGroups *MessageHandlerGroup
	err                  error
	exit                 chan bool
}

func (b *Bot) Alive() bool {
	if b.self == nil {
		return false
	}
	select {
	case <-b.exit:
		return false
	default:
		return true
	}
}

func (b *Bot) GetCurrentUser() (*Self, error) {
	if b.self == nil {
		return nil, errors.New("user not login")
	}
	return b.self, nil
}

func (b *Bot) Login() error {
	b.prepare()
	uuid, err := b.Caller.GetLoginUUID()
	if err != nil {
		return err
	}
	if b.UUIDCallback != nil {
		b.UUIDCallback(uuid)
	}
	for {
		resp, err := b.Caller.CheckLogin(uuid)
		if err != nil {
			return err
		}
		switch resp.Code {
		case statusSuccess:
			return b.login(resp.Raw)
		case statusScanned:
			if b.ScanCallBack != nil {
				b.ScanCallBack(resp.Raw)
			}
		case statusTimeout:
			return errors.New("login time out")
		case statusWait:
			continue
		}
	}
}

func (b *Bot) login(data []byte) error {
	if b.LoginCallBack != nil {
		b.LoginCallBack(data)
	}
	info, err := b.Caller.GetLoginInfo(data)
	if err != nil {
		return err
	}

	b.storage.SetLoginInfo(*info)

	request := BaseRequest{
		Uin:      info.WxUin,
		Sid:      info.WxSid,
		Skey:     info.SKey,
		DeviceID: GetRandomDeviceId(),
	}

	b.storage.SetBaseRequest(request)
	resp, err := b.Caller.WebInit(request)
	if err != nil {
		return err
	}
	b.self = &Self{Bot: b, User: &resp.User}
	b.storage.SetWebInitResponse(*resp)

	if err = b.Caller.WebWxStatusNotify(request, *resp, *info); err != nil {
		return err
	}
	go func() {
		b.stopAsyncCALL(b.asyncCall())
	}()
	return nil
}

func (b *Bot) asyncCall() error {
	var (
		err  error
		resp *SyncCheckResponse
	)
	for b.Alive() {
		info := b.storage.GetLoginInfo()
		response := b.storage.GetWebInitResponse()
		resp, err = b.Caller.SyncCheck(info, response)
		if err != nil {
			return err
		}
		if !resp.Success() {
			return fmt.Errorf("unknow code got %s", resp.RetCode)
		}
		if !resp.NorMal() {
			if err = b.getMessage(); err != nil {
				return err
			}
		}
	}
	return err
}

func (b *Bot) stopAsyncCALL(err error) {
	if err != nil {
		b.exit <- true
		b.err = err
	}
}
func (b *Bot) getMessage() error {
	info := b.storage.GetLoginInfo()
	response := b.storage.GetWebInitResponse()
	request := b.storage.GetBaseRequest()
	resp, err := b.Caller.WebWxSync(request, response, info)
	if err != nil {
		return err
	}
	response.SyncKey = resp.SyncKey
	b.storage.SetWebInitResponse(response)
	for _, message := range resp.AddMsgList {
		processMessage(message, b)
		b.messageHandlerGroups.ProcessMessage(message)
	}
	return nil
}

func (b *Bot) prepare() {
	if b.storage == nil {
		panic("WechatStorage can not be nil")
	}
	if b.messageHandlerGroups == nil {
		panic("message can not be nil")
	}
}

func (b *Bot) RegisterMessageHandler(handler MessageHandler) {
	b.messageHandlerGroups.RegisterHandler(handler)
}

func (b *Bot) Block() {
	<-b.exit
}

func NewBot(caller *Caller, storage WechatStorage) *Bot {
	return &Bot{Caller: caller, storage: storage}
}

func DefaultBot() *Bot {
	return NewBot(DefaultCaller(), NewSimpleWechatStorage())
}

func PrintlnQrcodeUrl(uuid string) {
	println(qrcodeUrl + uuid)
}
