package openwechat

// BotLogin 定义了一个Login的接口
type BotLogin interface {
	Login(bot *Bot) error
}

// SacnLogin 扫码登录
type SacnLogin struct{}

// Login 实现了 BotLogin 接口
func (s *SacnLogin) Login(bot *Bot) error {
	uuid, err := bot.Caller.GetLoginUUID()
	if err != nil {
		return err
	}
	return s.checkLogin(bot, uuid)
}

// checkLogin 该方法会一直阻塞，直到用户扫码登录，或者二维码过期
func (s *SacnLogin) checkLogin(bot *Bot, uuid string) error {
	bot.uuid = uuid
	bot.uuid = uuid
	loginChecker := &LoginChecker{
		Bot: bot,
		Tip: "0",
	}
	return loginChecker.CheckLogin()
}

type hotLoginOption struct {
	withRetry bool
	_         struct{}
}

type HotLoginOptionFunc func(o *hotLoginOption)

func HotLoginWithRetry(flag bool) HotLoginOptionFunc {
	return func(o *hotLoginOption) {
		o.withRetry = flag
	}
}

// HotLogin 热登录模式
type HotLogin struct {
	storage HotReloadStorage
	opt     hotLoginOption
}

// Login 实现了 BotLogin 接口
func (h *HotLogin) Login(bot *Bot) error {
	err := h.login(bot)
	if err != nil && h.opt.withRetry {
		scanLogin := SacnLogin{}
		return scanLogin.Login(bot)
	}
	return err
}

func (h *HotLogin) login(bot *Bot) error {
	if err := h.hotLoginInit(bot); err != nil {
		return err
	}
	return bot.WebInit()
}

func (h *HotLogin) hotLoginInit(bot *Bot) error {
	bot.hotReloadStorage = h.storage
	return bot.reload()
}

type pushLoginOption struct {
	withoutUUIDCallback  bool
	withoutScanCallback  bool
	withoutLoginCallback bool
	withRetry            bool
}

type PushLoginOptionFunc func(o *pushLoginOption)

// PushLoginWithoutUUIDCallback 设置 PushLogin 不执行二维码回调
func PushLoginWithoutUUIDCallback(flag bool) PushLoginOptionFunc {
	return func(o *pushLoginOption) {
		o.withoutUUIDCallback = flag
	}
}

// PushLoginWithoutScanCallback 设置 PushLogin 不执行扫码回调
func PushLoginWithoutScanCallback(flag bool) PushLoginOptionFunc {
	return func(o *pushLoginOption) {
		o.withoutScanCallback = flag
	}
}

// PushLoginWithoutLoginCallback 设置 PushLogin 不执行登录回调
func PushLoginWithoutLoginCallback(flag bool) PushLoginOptionFunc {
	return func(o *pushLoginOption) {
		o.withoutLoginCallback = flag
	}
}

// PushLoginWithRetry 设置 PushLogin 失败后执行扫码登录
func PushLoginWithRetry(flag bool) PushLoginOptionFunc {
	return func(o *pushLoginOption) {
		o.withRetry = flag
	}
}

// defaultPushLoginOpts 默认的 PushLogin
var defaultPushLoginOpts = [...]PushLoginOptionFunc{
	PushLoginWithoutUUIDCallback(true),
	PushLoginWithoutScanCallback(true),
}

// PushLogin 免扫码登录模式
type PushLogin struct {
	storage HotReloadStorage
	opt     pushLoginOption
}

// Login 实现了 BotLogin 接口
func (p PushLogin) Login(bot *Bot) error {
	if err := p.pushLoginInit(bot); err != nil {
		return err
	}
	resp, err := bot.Caller.WebWxPushLogin(bot.Storage.LoginInfo.WxUin)
	if err != nil {
		return err
	}
	if err = resp.Err(); err != nil {
		return err
	}
	err = p.checkLogin(bot, resp.UUID)
	if err != nil && p.opt.withRetry {
		scanLogin := SacnLogin{}
		return scanLogin.Login(bot)
	}
	return err
}

func (p PushLogin) pushLoginInit(bot *Bot) error {
	bot.hotReloadStorage = p.storage
	return bot.reload()
}

// checkLogin 登录检查
func (p PushLogin) checkLogin(bot *Bot, uuid string) error {
	bot.uuid = uuid
	loginChecker := &LoginChecker{
		Bot:                 bot,
		Tip:                 "1",
		WithLoginCallback:   p.opt.withoutLoginCallback,
		WithoutUUIDCallback: p.opt.withoutUUIDCallback,
		WithScanCallback:    p.opt.withoutScanCallback,
	}
	return loginChecker.CheckLogin()
}

type LoginChecker struct {
	Bot                 *Bot
	Tip                 string
	WithoutUUIDCallback bool
	WithLoginCallback   bool
	WithScanCallback    bool
}

func (l *LoginChecker) CheckLogin() error {
	uuid := l.Bot.UUID()
	// 二维码获取回调
	if l.Bot.UUIDCallback != nil && !l.WithoutUUIDCallback {
		l.Bot.UUIDCallback(uuid)
	}
	var tip = l.Tip
	for {
		// 长轮询检查是否扫码登录
		resp, err := l.Bot.Caller.CheckLogin(uuid, tip)
		if err != nil {
			return err
		}
		if tip == "1" {
			tip = "0"
		}
		switch resp.Code {
		case StatusSuccess:
			// 判断是否有登录回调，如果有执行它
			if err = l.Bot.HandleLogin(resp.Raw); err != nil {
				return err
			}
			if l.Bot.LoginCallBack != nil && !l.WithLoginCallback {
				l.Bot.LoginCallBack(resp.Raw)
			}
			return nil
		case StatusScanned:
			// 执行扫码回调
			if l.Bot.ScanCallBack != nil && !l.WithScanCallback {
				l.Bot.ScanCallBack(resp.Raw)
			}
		case StatusTimeout:
			return ErrLoginTimeout
		case StatusWait:
			continue
		}
	}
}
