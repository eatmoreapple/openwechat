package openwechat

import (
	"context"
	"encoding/json"
	"errors"
)

// LoginCode 定义登录状态码
type LoginCode string

const (
	// LoginCodeSuccess 登录成功
	LoginCodeSuccess LoginCode = "200"
	// LoginCodeScanned 已扫码
	LoginCodeScanned LoginCode = "201"
	// LoginCodeTimeout 登录超时
	LoginCodeTimeout LoginCode = "400"
	// LoginCodeWait 等待扫码
	LoginCodeWait LoginCode = "408"
)

func (l LoginCode) String() string {
	switch l {
	case LoginCodeSuccess:
		return "登录成功"
	case LoginCodeScanned:
		return "已扫码"
	case LoginCodeTimeout:
		return "登录超时"
	case LoginCodeWait:
		return "等待扫码"
	default:
		return "未知状态"
	}
}

type BotPreparer interface {
	Prepare(*Bot)
}

type BotLoginOption interface {
	BotPreparer
	OnError(*Bot, error) error
	OnSuccess(*Bot) error
}

// BotOptionGroup 是一个 BotLoginOption 的集合
// 用于将多个 BotLoginOption 组合成一个 BotLoginOption
type BotOptionGroup []BotLoginOption

// Prepare 实现了 BotLoginOption 接口
func (g BotOptionGroup) Prepare(bot *Bot) {
	for _, option := range g {
		option.Prepare(bot)
	}
}

// OnError 实现了 BotLoginOption 接口
func (g BotOptionGroup) OnError(b *Bot, err error) error {
	// 当有一个 BotLoginOption 的 OnError 返回的 error 等于 nil 时，就会停止执行后续的 BotLoginOption
	for _, option := range g {
		currentErr := option.OnError(b, err)
		if currentErr == nil {
			return nil
		}
		if currentErr != err {
			return currentErr
		}
	}
	return err
}

// OnSuccess 实现了 BotLoginOption 接口
func (g BotOptionGroup) OnSuccess(b *Bot) error {
	for _, option := range g {
		if err := option.OnSuccess(b); err != nil {
			return err
		}
	}
	return nil
}

type BaseBotLoginOption struct{}

func (BaseBotLoginOption) Prepare(_ *Bot) {}

func (BaseBotLoginOption) OnError(_ *Bot, err error) error { return err }

func (BaseBotLoginOption) OnSuccess(_ *Bot) error { return nil }

// DoNothingBotLoginOption 是一个空的 BotLoginOption，表示不做任何操作
var DoNothingBotLoginOption = &BaseBotLoginOption{}

// RetryLoginOption 在登录失败后进行扫码登录
type RetryLoginOption struct {
	BaseBotLoginOption
	MaxRetryCount    int
	currentRetryTime int
}

// OnError 实现了 BotLoginOption 接口
// 当登录失败后，会调用此方法进行扫码登录
func (r *RetryLoginOption) OnError(bot *Bot, err error) error {
	if r.currentRetryTime >= r.MaxRetryCount {
		return err
	}
	r.currentRetryTime++
	return bot.Login()
}

func NewRetryLoginOption() BotLoginOption {
	return &RetryLoginOption{MaxRetryCount: 1}
}

type BotPreparerFunc func(*Bot)

func (f BotPreparerFunc) Prepare(b *Bot) {
	f(b)
}

// withMode 是一个 BotPreparerFunc，用于设置 Bot 的模式
func withMode(mode Mode) BotPreparer {
	return BotPreparerFunc(func(b *Bot) { b.Caller.Client.SetMode(mode) })
}

// btw, 这两个变量已经变了4回了, 但是为了兼容以前的代码, 还是得想着法儿让用户无感知的更新
var (
	// Normal 网页版微信模式
	Normal = withMode(normal)

	// Desktop 桌面微信模式
	Desktop = withMode(desktop)
)

// WithContextOption 是一个 BotPreparerFunc，用于设置 Bot 的 context
func WithContextOption(ctx context.Context) BotPreparer {
	if ctx == nil {
		panic("context is nil")
	}
	return BotPreparerFunc(func(b *Bot) { b.context, b.cancel = context.WithCancel(ctx) })
}

// WithUUIDOption 是一个 BotPreparerFunc，用于设置 Bot 的 登录 uuid
func WithUUIDOption(uuid string) BotPreparer {
	return BotPreparerFunc(func(b *Bot) { b.loginUUID = uuid })
}

// WithDeviceID 是一个 BotPreparerFunc，用于设置 Bot 的 设备 id
func WithDeviceID(deviceId string) BotPreparer {
	return BotPreparerFunc(func(b *Bot) { b.deviceId = deviceId })
}

// BotLogin 定义了一个Login的接口
type BotLogin interface {
	Login(bot *Bot) error
}

// ScanLogin 扫码登录
type ScanLogin struct {
	UUID string
}

// Login 实现了 BotLogin 接口
func (s *ScanLogin) Login(bot *Bot) error {
	var uuid = s.UUID
	if uuid == "" {
		var err error
		uuid, err = bot.Caller.GetLoginUUID(bot.Context())
		if err != nil {
			return err
		}
	}
	return s.checkLogin(bot, uuid)
}

// checkLogin 该方法会一直阻塞，直到用户扫码登录，或者二维码过期
func (s *ScanLogin) checkLogin(bot *Bot, uuid string) error {
	bot.uuid = uuid
	loginChecker := &LoginChecker{
		Bot:           bot,
		Tip:           "0",
		UUIDCallback:  bot.UUIDCallback,
		LoginCallBack: bot.LoginCallBack,
		ScanCallBack:  bot.ScanCallBack,
	}
	return loginChecker.CheckLogin()
}

func botReload(bot *Bot, storage HotReloadStorage) error {
	if storage == nil {
		return errors.New("storage is nil")
	}
	bot.hotReloadStorage = storage
	var item HotReloadStorageItem
	if err := json.NewDecoder(storage).Decode(&item); err != nil {
		return err
	}
	bot.Caller.Client.SetCookieJar(item.Jar)
	bot.Storage.LoginInfo = item.LoginInfo
	bot.Storage.Request = item.BaseRequest
	bot.Caller.Client.Domain = item.WechatDomain
	bot.uuid = item.UUID
	if item.SyncKey != nil {
		if bot.Storage.Response == nil {
			bot.Storage.Response = &WebInitResponse{}
		}
		bot.Storage.Response.SyncKey = item.SyncKey
	}
	return nil
}

// HotLogin 热登录模式
type HotLogin struct {
	storage HotReloadStorage
}

// Login 实现了 BotLogin 接口
func (h *HotLogin) Login(bot *Bot) error {
	if err := botReload(bot, h.storage); err != nil {
		return err
	}
	return bot.webInit()
}

// PushLogin 免扫码登录模式
type PushLogin struct {
	storage HotReloadStorage
}

// Login 实现了 BotLogin 接口
func (p *PushLogin) Login(bot *Bot) error {
	if err := botReload(bot, p.storage); err != nil {
		return err
	}
	resp, err := bot.Caller.WebWxPushLogin(bot.Context(), bot.Storage.LoginInfo.WxUin)
	if err != nil {
		return err
	}
	if err = resp.Err(); err != nil {
		return err
	}
	return p.checkLogin(bot, resp.UUID)
}

// checkLogin 登录检查
func (p *PushLogin) checkLogin(bot *Bot, uuid string) error {
	bot.uuid = uuid
	// 为什么把 UUIDCallback 和 ScanCallBack 置为nil呢?
	// 因为这两个对用户是无感知的。
	loginChecker := &LoginChecker{
		Bot:           bot,
		Tip:           "1",
		LoginCallBack: bot.LoginCallBack,
	}
	return loginChecker.CheckLogin()
}

type LoginChecker struct {
	Bot           *Bot
	Tip           string
	UUIDCallback  func(uuid string)
	LoginCallBack func(body CheckLoginResponse)
	ScanCallBack  func(body CheckLoginResponse)
}

func (l *LoginChecker) CheckLogin() error {
	uuid := l.Bot.UUID()
	// 二维码获取回调
	if cb := l.UUIDCallback; cb != nil {
		cb(uuid)
	}
	var tip = l.Tip
	for {
		// 长轮询检查是否扫码登录
		resp, err := l.Bot.Caller.CheckLogin(l.Bot.Context(), uuid, tip)
		if err != nil {
			return err
		}
		code, err := resp.Code()
		if err != nil {
			return err
		}
		if tip == "1" {
			tip = "0"
		}
		switch code {
		case LoginCodeSuccess:
			// 判断是否有登录回调，如果有执行它
			redirectURL, err := resp.RedirectURL()
			if err != nil {
				return err
			}
			if err = l.Bot.loginFromURL(redirectURL); err != nil {
				return err
			}
			if cb := l.LoginCallBack; cb != nil {
				cb(resp)
			}
			return nil
		case LoginCodeScanned:
			// 执行扫码回调
			if cb := l.ScanCallBack; cb != nil {
				cb(resp)
			}
		case LoginCodeTimeout:
			return ErrLoginTimeout
		case LoginCodeWait:
			continue
		}
	}
}

// # 下面都是即将废弃的函数。
// # 为了兼容老版本暂时留了下来, 但是它的函数签名已经发生了改变。
// # 如果你是使用的是openwechat提供的api来调用这些函数，那么你是感知不到变动的。
// # openwechat内部对这些函数的调用做了兼容处理, 如果你的代码中调用了这些函数, 请尽快修改。

// Deprecated: 请使用 NewRetryLoginOption 代替
// HotLoginWithRetry 热登录模式，如果登录失败会重试
func HotLoginWithRetry(flag bool) BotLoginOption {
	if flag {
		return NewRetryLoginOption()
	}
	return DoNothingBotLoginOption
}

// Deprecated: 请使用 NewRetryLoginOption 代替
// PushLoginWithRetry 免扫码登录模式，如果登录失败会重试
func PushLoginWithRetry(flag bool) BotLoginOption {
	if !flag {
		return DoNothingBotLoginOption
	}
	return NewRetryLoginOption()
}
