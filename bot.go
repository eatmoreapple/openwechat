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

// 判断当前用户是否正常在线
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

// 获取当前的用户
func (b *Bot) GetCurrentUser() (*Self, error) {
	if b.self == nil {
		return nil, errors.New("user not login")
	}
	return b.self, nil
}

// 用户登录
// 该方法会一直阻塞，直到用户扫码登录，或者二维码过期
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

// 登录逻辑
func (b *Bot) login(data []byte) error {
	// 判断是否有登录回调，如果有执行它
	if b.LoginCallBack != nil {
		b.LoginCallBack(data)
	}
	// 获取登录的一些基本的信息
	info, err := b.Caller.GetLoginInfo(data)
	if err != nil {
		return err
	}

	// 将LoginInfo存到storage里面
	b.storage.SetLoginInfo(*info)

	// 构建BaseRequest
	request := BaseRequest{
		Uin:      info.WxUin,
		Sid:      info.WxSid,
		Skey:     info.SKey,
		DeviceID: GetRandomDeviceId(),
	}

	// 将BaseRequest存到storage里面方便后续调用
	b.storage.SetBaseRequest(request)
	// 获取初始化的用户信息和一些必要的参数
	resp, err := b.Caller.WebInit(request)
	if err != nil {
		return err
	}
	// 设置当前的用户
	b.self = &Self{Bot: b, User: &resp.User}
	b.self.Self = b.self
	b.storage.SetWebInitResponse(*resp)

	// 通知手机客户端已经登录
	if err = b.Caller.WebWxStatusNotify(request, *resp, *info); err != nil {
		return err
	}
	// 开启协程，轮训获取是否有新的消息返回
	go func() {
		b.stopAsyncCALL(b.asyncCall())
	}()
	return nil
}

// 轮训请求
// 根据状态码判断是否有新的请求
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
		// 如果不是正常的状态码返回，发生了错误，直接退出
		if !resp.Success() {
			return fmt.Errorf("unknow code got %s", resp.RetCode)
		}
		// 如果Selector不为0，则获取消息
		if !resp.NorMal() {
			if err = b.getMessage(); err != nil {
				return err
			}
		}
	}
	return err
}

func (b *Bot) stopAsyncCALL(err error) {
	b.exit <- true
	b.err = err
	b.self = nil
}

// 获取新的消息
func (b *Bot) getMessage() error {
	info := b.storage.GetLoginInfo()
	response := b.storage.GetWebInitResponse()
	request := b.storage.GetBaseRequest()
	resp, err := b.Caller.WebWxSync(request, response, info)
	if err != nil {
		return err
	}
	// 更新SyncKey并且重新存入storage
	response.SyncKey = resp.SyncKey
	b.storage.SetWebInitResponse(response)
	// 遍历所有的新的消息，依次处理
	for _, message := range resp.AddMsgList {
		// 根据不同的消息类型来进行处理，方便后续统一调用
		processMessage(message, b)
		// 调用自定义的处理方法
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

// 注册消息处理的函数
func (b *Bot) RegisterMessageHandler(handler MessageHandler) {
	if b.messageHandlerGroups == nil {
		b.messageHandlerGroups = &MessageHandlerGroup{}
	}
	b.messageHandlerGroups.RegisterHandler(handler)
}

// 当消息同步发生了错误或者用户主动在手机上退出，该方法会立即返回，否则会一直阻塞
func (b *Bot) Block() error {
	if b.self == nil {
		return errors.New("`Block` must be called after user login")
	}
	<-b.exit
	return nil
}

func NewBot(caller *Caller, storage WechatStorage) *Bot {
	return &Bot{Caller: caller, storage: storage, exit: make(chan bool)}
}

func DefaultBot() *Bot {
	return NewBot(DefaultCaller(), NewSimpleWechatStorage())
}

func GetQrcodeUrl(uuid string) string {
	return qrcodeUrl + uuid
}

func PrintlnQrcodeUrl(uuid string) {
	println("访问下面网址扫描二维码登录")
	println(GetQrcodeUrl(uuid))
}
