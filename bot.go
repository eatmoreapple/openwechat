package openwechat

import (
    "errors"
    "log"
    "net/url"
)

type Bot struct {
    ScanCallBack     func(body []byte)
    LoginCallBack    func(body []byte)
    UUIDCallback     func(uuid string)
    MessageHandler   func(msg *Message)
    isHot            bool
    err              error
    exit             chan bool
    Caller           *Caller
    self             *Self
    storage          *Storage
    hotReloadStorage HotReloadStorage
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

func (b *Bot) HotLogin(storage HotReloadStorage, retry ...bool) error {
    b.isHot = true
    b.hotReloadStorage = storage

    var err error

    // 如果load出错了,就执行正常登陆逻辑
    // 第一次没有数据load都会出错的
    if err = storage.Load(); err != nil {
        return b.Login()
    }

    if err = b.hotLoginInit(); err != nil {
        return err
    }

    // 如果webInit出错,则说明可能身份信息已经失效
    // 如果retry为True的话,则进行正常登陆
    if err = b.webInit(); err != nil {
        if len(retry) > 0 {
            if retry[0] {
                return b.Login()
            }
        }
    }
    return err
}

// 热登陆初始化
func (b *Bot) hotLoginInit() error {
    cookies := b.hotReloadStorage.GetCookie()
    for u, ck := range cookies {
        path, err := url.Parse(u)
        if err != nil {
            return err
        }
        b.Caller.Client.Jar.SetCookies(path, ck)
    }
    b.storage.LoginInfo = b.hotReloadStorage.GetLoginInfo()
    b.storage.Request = b.hotReloadStorage.GetBaseRequest()
    return nil
}

// 用户登录
// 该方法会一直阻塞，直到用户扫码登录，或者二维码过期
func (b *Bot) Login() error {
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
            return b.handleLogin(resp.Raw)
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

func (b *Bot) Logout() error {
    if b.Alive() {
        info := b.storage.LoginInfo
        if err := b.Caller.Logout(info); err != nil {
            return err
        }
        b.stopAsyncCALL(errors.New("logout"))
        return nil
    }
    return errors.New("user not login")
}

// 登录逻辑
func (b *Bot) handleLogin(data []byte) error {
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
    b.storage.LoginInfo = info

    // 构建BaseRequest
    request := &BaseRequest{
        Uin:      info.WxUin,
        Sid:      info.WxSid,
        Skey:     info.SKey,
        DeviceID: GetRandomDeviceId(),
    }

    // 将BaseRequest存到storage里面方便后续调用
    b.storage.Request = request

    // 如果是热登陆,则将当前的重要信息写入hotReloadStorage
    if b.isHot {
        cookies := b.Caller.Client.GetCookieMap()
        if err := b.hotReloadStorage.Dump(cookies, request, info); err != nil {
            return err
        }
    }

    return b.webInit()
}

func (b *Bot) webInit() error {
    req := b.storage.Request
    info := b.storage.LoginInfo
    // 获取初始化的用户信息和一些必要的参数
    resp, err := b.Caller.WebInit(req)
    if err != nil {
        return err
    }
    // 设置当前的用户
    b.self = &Self{Bot: b, User: &resp.User}
    b.self.Self = b.self
    b.storage.Response = resp

    // 通知手机客户端已经登录
    if err = b.Caller.WebWxStatusNotify(req, resp, info); err != nil {
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
        // 长轮训检查是否有消息返回
        resp, err = b.Caller.SyncCheck(b.storage.LoginInfo, b.storage.Response)
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

func (b *Bot) stopAsyncCALL(err error) {
    b.exit <- true
    b.err = err
    b.self = nil
    log.Printf("exit with : %s", err.Error())
}

// 获取新的消息
func (b *Bot) getNewWechatMessage() error {
    resp, err := b.Caller.WebWxSync(b.storage.Request, b.storage.Response, b.storage.LoginInfo)
    if err != nil {
        return err
    }
    // 更新SyncKey并且重新存入storage
    b.storage.Response.SyncKey = resp.SyncKey
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

// 当消息同步发生了错误或者用户主动在手机上退出，该方法会立即返回，否则会一直阻塞
func (b *Bot) Block() error {
    if b.self == nil {
        return errors.New("`Block` must be called after user login")
    }
    if _, closed := <-b.exit; !closed {
        return errors.New("can not call `Block` after user logout")
    }
    close(b.exit)
    return nil
}

// 获取当前Bot崩溃的原因
func (b *Bot) CrashReason() error {
    return b.err
}

func NewBot(caller *Caller) *Bot {
    return &Bot{Caller: caller, storage: &Storage{}, exit: make(chan bool)}
}

func DefaultBot(modes ...mode) *Bot {
    var m mode
    if len(modes) == 0 {
        m = Normal
    } else {
        m = modes[0]
    }
    urlManager := GetUrlManagerByMode(m)
    return NewBot(DefaultCaller(urlManager))
}

func GetQrcodeUrl(uuid string) string {
    return qrcodeUrl + uuid
}

func PrintlnQrcodeUrl(uuid string) {
    println("访问下面网址扫描二维码登录")
    println(GetQrcodeUrl(uuid))
}
