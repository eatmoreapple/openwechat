package openwechat

import (
    "context"
    "encoding/json"
    "errors"
    "log"
    "net/url"
    "sync"
)

type Bot struct {
    ScanCallBack           func(body []byte) // 扫码回调,可获取扫码用户的头像
    LoginCallBack          func(body []byte) // 登陆回调
    LogoutCallBack         func(bot *Bot)    // 退出回调
    UUIDCallback           func(uuid string) // 获取UUID的回调函数
    MessageHandler         MessageHandler    // 获取消息成功的handle
    GetMessageErrorHandler func(err error)   // 获取消息发生错误的handle
    IsHot                  bool              // 是否为热登录模式
    once                   sync.Once
    err                    error
    context                context.Context
    cancel                 context.CancelFunc
    Caller                 *Caller
    self                   *Self
    Storage                *Storage
    HotReloadStorage       HotReloadStorage
}

// Alive 判断当前用户是否正常在线
func (b *Bot) Alive() bool {
    if b.self == nil {
        return false
    }
    select {
    case <-b.context.Done():
        return false
    default:
        return true
    }
}

// GetCurrentUser 获取当前的用户
//		self, err := bot.GetCurrentUser()
//		if err != nil {
//			return
//		}
//		fmt.Println(self.NickName)
func (b *Bot) GetCurrentUser() (*Self, error) {
    if b.self == nil {
        return nil, errors.New("user not login")
    }
    return b.self, nil
}

// HotLogin 热登录,可实现重复登录,
// retry设置为true可在热登录失效后进行普通登录行为
//		Storage := NewJsonFileHotReloadStorage("Storage.json")
//		err := bot.HotLogin(Storage, true)
//		fmt.Println(err)
func (b *Bot) HotLogin(storage HotReloadStorage, retry ...bool) error {
    b.IsHot = true
    b.HotReloadStorage = storage

    var err error

    // 如果load出错了,就执行正常登陆逻辑
    // 第一次没有数据load都会出错的
    item, err := NewHotReloadStorageItem(storage)

    if err != nil {
        return b.Login()
    }

    defer storage.Close()

    if err = b.hotLoginInit(*item); err != nil {
        return err
    }

    // 如果webInit出错,则说明可能身份信息已经失效
    // 如果retry为True的话,则进行正常登陆
    if err = b.WebInit(); err != nil {
        if len(retry) > 0 && retry[0] {
            return b.Login()
        }
    }
    return err
}

// 热登陆初始化
func (b *Bot) hotLoginInit(item HotReloadStorageItem) error {
    cookies := item.Cookies
    for u, ck := range cookies {
        path, err := url.Parse(u)
        if err != nil {
            return err
        }
        b.Caller.Client.Jar.SetCookies(path, ck)
    }
    b.Storage.LoginInfo = item.LoginInfo
    b.Storage.Request = item.BaseRequest
    b.Caller.Client.Domain = item.WechatDomain
    return nil
}

// Login 用户登录
// 该方法会一直阻塞，直到用户扫码登录，或者二维码过期
func (b *Bot) Login() error {
    uuid, err := b.Caller.GetLoginUUID()
    if err != nil {
        return err
    }
    // 二维码获取回调
    if b.UUIDCallback != nil {
        b.UUIDCallback(uuid)
    }
    for {
        // 长轮询检查是否扫码登录
        resp, err := b.Caller.CheckLogin(uuid)
        if err != nil {
            return err
        }
        switch resp.Code {
        case StatusSuccess:
            // 判断是否有登录回调，如果有执行它
            if b.LoginCallBack != nil {
                b.LoginCallBack(resp.Raw)
            }
            return b.HandleLogin(resp.Raw)
        case StatusScanned:
            // 执行扫码回调
            if b.ScanCallBack != nil {
                b.ScanCallBack(resp.Raw)
            }
        case StatusTimeout:
            return ErrLoginTimeout
        case StatusWait:
            continue
        }
    }
}

// Logout 用户退出
func (b *Bot) Logout() error {
    if b.Alive() {
        if b.LogoutCallBack != nil {
            b.LogoutCallBack(b)
        }
        info := b.Storage.LoginInfo
        if err := b.Caller.Logout(info); err != nil {
            return err
        }
        b.stopAsyncCALL(errors.New("logout"))
        return nil
    }
    return errors.New("user not login")
}

// HandleLogin 登录逻辑
func (b *Bot) HandleLogin(data []byte) error {
    // 获取登录的一些基本的信息
    info, err := b.Caller.GetLoginInfo(data)
    if err != nil {
        return err
    }
    // 将LoginInfo存到storage里面
    b.Storage.LoginInfo = info

    // 构建BaseRequest
    request := &BaseRequest{
        Uin:      info.WxUin,
        Sid:      info.WxSid,
        Skey:     info.SKey,
        DeviceID: GetRandomDeviceId(),
    }

    // 将BaseRequest存到storage里面方便后续调用
    b.Storage.Request = request

    // 如果是热登陆,则将当前的重要信息写入hotReloadStorage
    if b.IsHot {
        if err = b.DumpHotReloadStorage(); err != nil {
            return err
        }
    }

    return b.WebInit()
}

// WebInit 根据有效凭证获取和初始化用户信息
func (b *Bot) WebInit() error {
    req := b.Storage.Request
    info := b.Storage.LoginInfo
    // 获取初始化的用户信息和一些必要的参数
    resp, err := b.Caller.WebInit(req)
    if err != nil {
        return err
    }
    // 设置当前的用户
    b.self = &Self{Bot: b, User: &resp.User}
    b.self.Self = b.self
    b.Storage.Response = resp

    // 通知手机客户端已经登录
    if err = b.Caller.WebWxStatusNotify(req, resp, info); err != nil {
        return err
    }
    // 开启协程，轮询获取是否有新的消息返回

    // FIX: 当bot在线的情况下执行热登录,会开启多次事件监听
    go b.once.Do(func() {
        if b.GetMessageErrorHandler == nil {
            b.GetMessageErrorHandler = b.stopAsyncCALL
        }
        if err = b.asyncCall(); err != nil {
            b.GetMessageErrorHandler(err)
        }
    })
    return nil
}

// 轮询请求
// 根据状态码判断是否有新的请求
func (b *Bot) asyncCall() error {
    var (
        err  error
        resp *SyncCheckResponse
    )
    for b.Alive() {
        // 长轮询检查是否有消息返回
        resp, err = b.Caller.SyncCheck(b.Storage.LoginInfo, b.Storage.Response)
        if err != nil {
            return err
        }
        // 如果不是正常的状态码返回，发生了错误，直接退出
        if !resp.Success() {
            return resp
        }
        // 如果Selector不为0，则获取消息
        if !resp.NorMal() {
            if err = b.getNewWechatMessage(); err != nil {
                return err
            }
        }
    }
    return err
}

// 当获取消息发生错误时, 默认的错误处理行为
func (b *Bot) stopAsyncCALL(err error) {
    b.cancel()
    b.err = err
    b.self = nil
    log.Printf("exit with : %s", err.Error())
}

// 获取新的消息
func (b *Bot) getNewWechatMessage() error {
    resp, err := b.Caller.WebWxSync(b.Storage.Request, b.Storage.Response, b.Storage.LoginInfo)
    if err != nil {
        return err
    }
    // 更新SyncKey并且重新存入storage
    b.Storage.Response.SyncKey = resp.SyncKey
    // 遍历所有的新的消息，依次处理
    for _, message := range resp.AddMsgList {
        // 根据不同的消息类型来进行处理，方便后续统一调用
        message.init(b)
        // 调用自定义的处理方法
        if handler := b.MessageHandler; handler != nil {
            handler(message)
        }
    }
    return nil
}

// Block 当消息同步发生了错误或者用户主动在手机上退出，该方法会立即返回，否则会一直阻塞
func (b *Bot) Block() error {
    if b.self == nil {
        return errors.New("`Block` must be called after user login")
    }
    <-b.context.Done()
    return nil
}

// CrashReason 获取当前Bot崩溃的原因
func (b *Bot) CrashReason() error {
    return b.err
}

// MessageOnSuccess setter for Bot.MessageHandler
func (b *Bot) MessageOnSuccess(h func(msg *Message)) {
    b.MessageHandler = h
}

// MessageOnError setter for Bot.GetMessageErrorHandler
func (b *Bot) MessageOnError(h func(err error)) {
    b.GetMessageErrorHandler = h
}

// DumpHotReloadStorage 写入HotReloadStorage
func (b *Bot) DumpHotReloadStorage() error {
    if b.HotReloadStorage == nil {
        return errors.New("HotReloadStorage can not be nil")
    }
    cookies := b.Caller.Client.GetCookieMap()
    item := HotReloadStorageItem{
        BaseRequest:  b.Storage.Request,
        Cookies:      cookies,
        LoginInfo:    b.Storage.LoginInfo,
        WechatDomain: b.Caller.Client.Domain,
    }

    data, err := json.Marshal(item)
    if err != nil {
        return err
    }
    if _, err = b.HotReloadStorage.Write(data); err != nil {
        return err
    }
    return b.HotReloadStorage.Close()
}

// OnLogin is a setter for LoginCallBack
func (b *Bot) OnLogin(f func(body []byte)) {
    b.LoginCallBack = f
}

// OnScanned is a setter for ScanCallBack
func (b *Bot) OnScanned(f func(body []byte)) {
    b.ScanCallBack = f
}

// OnLogout is a setter for LogoutCallBack
func (b *Bot) OnLogout(f func(bot *Bot)) {
    b.LogoutCallBack = f
}

// NewBot Bot的构造方法，需要自己传入Caller
func NewBot(caller *Caller) *Bot {
    ctx, cancel := context.WithCancel(context.Background())
    return &Bot{Caller: caller, Storage: &Storage{}, context: ctx, cancel: cancel}
}

// DefaultBot 默认的Bot的构造方法,
// mode不传入默认为openwechat.Normal,详情见mode
//     bot := openwechat.DefaultBot(openwechat.Desktop)
func DefaultBot(modes ...mode) *Bot {
    var m mode
    if len(modes) == 0 {
        m = Normal
    } else {
        m = modes[0]
    }
    caller := DefaultCaller()
    caller.Client.mode = m
    bot := NewBot(caller)
    bot.UUIDCallback = PrintlnQrcodeUrl
    return bot
}

// GetQrcodeUrl 通过uuid获取登录二维码的url
func GetQrcodeUrl(uuid string) string {
    return qrcode + uuid
}

// PrintlnQrcodeUrl 打印登录二维码
func PrintlnQrcodeUrl(uuid string) {
    println("访问下面网址扫描二维码登录")
    println(GetQrcodeUrl(uuid))
}
