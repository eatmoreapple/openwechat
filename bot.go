package openwechat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os/exec"
	"runtime"
)

type Bot struct {
	ScanCallBack        func(body CheckLoginResponse) // 扫码回调,可获取扫码用户的头像
	LoginCallBack       func(body CheckLoginResponse) // 登陆回调
	LogoutCallBack      func(bot *Bot)                // 退出回调
	UUIDCallback        func(uuid string)             // 获取UUID的回调函数
	SyncCheckCallback   func(resp SyncCheckResponse)  // 心跳回调
	MessageHandler      MessageHandler                // 获取消息成功的handle
	MessageErrorHandler MessageErrorHandler           // 获取消息发生错误的handle, 返回err == nil 则尝试继续监听
	Caller              *Caller
	Storage             *Session
	err                 error
	context             context.Context
	cancel              func()
	self                *Self
	hotReloadStorage    HotReloadStorage
	uuid                string
	loginUUID           string
	deviceId            string // 设备Id
	loginOptionGroup    BotOptionGroup
}

// Alive 判断当前用户是否正常在线
func (b *Bot) Alive() bool {
	select {
	case <-b.context.Done():
		return false
	default:
		return b.self != nil
	}
}

// GetCurrentUser 获取当前的用户
//
//	self, err := bot.GetCurrentUser()
//	if err != nil {
//		return
//	}
//	fmt.Println(self.NickName)
func (b *Bot) GetCurrentUser() (*Self, error) {
	if b.self == nil {
		return nil, errors.New("user not login")
	}
	return b.self, nil
}

// login 这里对进行一些对登录前后的hook
func (b *Bot) login(login BotLogin) (err error) {
	opt := b.loginOptionGroup
	opt.Prepare(b)
	if err = login.Login(b); err != nil {
		err = opt.OnError(b, err)
	}
	if err != nil {
		return err
	}
	return opt.OnSuccess(b)
}

// Login 用户登录
func (b *Bot) Login() error {
	scanLogin := &ScanLogin{UUID: b.loginUUID}
	return b.login(scanLogin)
}

// HotLogin 热登录,可实现在单位时间内免重复扫码登录
// 热登录需要先扫码登录一次才可以进行热登录
func (b *Bot) HotLogin(storage HotReloadStorage, opts ...BotLoginOption) error {
	hotLogin := &HotLogin{storage: storage}
	// 进行相关设置。
	// 如果相对默认的行为进行修改，在opts里面进行追加即可。
	b.loginOptionGroup = opts
	return b.login(hotLogin)
}

// PushLogin 免扫码登录
// 免扫码登录需要先扫码登录一次才可以进行扫码登录
func (b *Bot) PushLogin(storage HotReloadStorage, opts ...BotLoginOption) error {
	pushLogin := &PushLogin{storage: storage}
	// 进行相关设置。
	// 如果相对默认的行为进行修改，在opts里面进行追加即可。
	b.loginOptionGroup = opts
	return b.login(pushLogin)
}

// Logout 用户退出
func (b *Bot) Logout() error {
	if b.Alive() {
		info := b.Storage.LoginInfo
		if err := b.Caller.Logout(b.Context(), info); err != nil {
			return err
		}
		b.ExitWith(ErrUserLogout)
		return nil
	}
	return errors.New("user not login")
}

// loginFromURL 登录逻辑
func (b *Bot) loginFromURL(path *url.URL) error {
	// 获取登录的一些基本的信息
	info, err := b.Caller.GetLoginInfo(b.Context(), path)
	if err != nil {
		return err
	}
	// 将LoginInfo存到storage里面
	b.Storage.LoginInfo = info

	// 处理设备Id
	if b.deviceId == "" {
		b.deviceId = GetRandomDeviceId()
	}

	// 构建BaseRequest
	request := &BaseRequest{
		Uin:      info.WxUin,
		Sid:      info.WxSid,
		Skey:     info.SKey,
		DeviceID: b.deviceId,
	}

	// 将BaseRequest存到storage里面方便后续调用
	b.Storage.Request = request

	return b.webInit()
}

func (b *Bot) initContacts(resp *WebInitResponse) {
	if resp.ContactList != nil {
		resp.ContactList.init(b.self)
	}
}

func (b *Bot) updateSyncKey(resp *WebInitResponse) {
	if b.Storage.Response != nil {
		resp.SyncKey = b.Storage.Response.SyncKey
	}
	b.Storage.Response = resp
}

// 拆分为小函数
func (b *Bot) initUserInfo() error {
	resp, err := b.Caller.WebInit(b.Context(), b.Storage.Request)
	if err != nil {
		return err
	}

	b.initSelf(resp)
	b.initContacts(resp)
	b.updateSyncKey(resp)

	return nil
}

func (b *Bot) initSelf(resp *WebInitResponse) {
	b.self = &Self{
		bot:  b,
		User: resp.User,
	}
	b.self.formatEmoji()
	b.self.self = b.self
}

func (b *Bot) saveHotReloadData() error {
	if b.hotReloadStorage == nil {
		return nil
	}
	return b.DumpHotReloadStorage()
}

func (b *Bot) notifyMobileClient() error {
	notifyOption := &CallerWebWxStatusNotifyOptions{
		BaseRequest:     b.Storage.Request,
		WebInitResponse: b.Storage.Response,
		LoginInfo:       b.Storage.LoginInfo,
	}
	return b.Caller.WebWxStatusNotify(b.Context(), notifyOption)
}

func (b *Bot) startMessageSync() {
	go b.runMessageLoop()
}

func (b *Bot) runMessageLoop() {
	b.initMessageErrorHandler()

	for b.Alive() {
		if err := b.syncCheck(); err != nil {
			if err = b.handleSyncError(err); err != nil {
				b.ExitWith(err)
				return
			}
		}
	}
}

func (b *Bot) initMessageErrorHandler() {
	if b.MessageErrorHandler == nil {
		b.MessageErrorHandler = defaultMessageErrorHandler
	}
}

func (b *Bot) handleSyncError(err error) error {
	return b.MessageErrorHandler(err)
}

func (b *Bot) webInit() error {
	// 1. 初始化用户信息
	if err := b.initUserInfo(); err != nil {
		return fmt.Errorf("init user info: %w", err)
	}

	// 2. 保存热重载数据
	if err := b.saveHotReloadData(); err != nil {
		return fmt.Errorf("save hot reload data: %w", err)
	}

	// 3. 通知移动端
	if err := b.notifyMobileClient(); err != nil {
		return fmt.Errorf("notify mobile client: %w", err)
	}

	// 4. 启动消息同步
	b.startMessageSync()

	return nil
}

func (b *Bot) executeSyncCallback(resp *SyncCheckResponse) {
	if b.SyncCheckCallback != nil {
		b.SyncCheckCallback(*resp)
	}
}

func (b *Bot) handleMessages(messages []*Message) {
	if b.MessageHandler == nil {
		return
	}

	for _, msg := range messages {
		msg.init(b)
		b.MessageHandler(msg)
	}
}

func (b *Bot) processNewMessages() error {
	// 获取新消息
	messages, err := b.syncMessage()
	if err != nil {
		return fmt.Errorf("sync message failed: %w", err)
	}

	// 保存热重载数据
	_ = b.DumpHotReloadStorage()

	// 处理消息
	b.handleMessages(messages)

	return nil
}

func (b *Bot) handleSyncSelector(selector Selector) error {
	switch selector {
	case SelectorNormal:
		return nil
	default:
		return b.processNewMessages()
	}
}

func (b *Bot) doSyncCheck(option *CallerSyncCheckOptions) error {
	// 更新同步检查参数
	b.updateSyncCheckOptions(option)

	// 执行同步检查
	resp, err := b.Caller.SyncCheck(b.Context(), option)
	if err != nil {
		return fmt.Errorf("sync check failed: %w", err)
	}

	// 执行心跳回调
	b.executeSyncCallback(resp)

	// 检查响应状态
	if err := resp.Err(); err != nil {
		return resp.Err()
	}

	// 处理消息
	return b.handleSyncSelector(resp.Selector)
}

func (b *Bot) updateSyncCheckOptions(option *CallerSyncCheckOptions) {
	option.BaseRequest = b.Storage.Request
	option.WebInitResponse = b.Storage.Response
	option.LoginInfo = b.Storage.LoginInfo
}

// 轮询请求
// 根据状态码判断是否有新的请求
func (b *Bot) syncCheck() error {
	option := &CallerSyncCheckOptions{}

	for b.Alive() {
		if err := b.doSyncCheck(option); err != nil {
			return err
		}
	}
	return nil
}

// 获取新的消息
func (b *Bot) syncMessage() ([]*Message, error) {
	opt := CallerWebWxSyncOptions{
		BaseRequest:     b.Storage.Request,
		WebInitResponse: b.Storage.Response,
		LoginInfo:       b.Storage.LoginInfo,
	}
	resp, err := b.Caller.WebWxSync(b.Context(), &opt)
	if err != nil {
		return nil, err
	}

	// 更新SyncKey并且重新存入storage 如获取到的SyncKey为空则不更新
	if resp.SyncKey.Count > 0 {
		b.Storage.Response.SyncKey = resp.SyncKey
	}
	return resp.AddMsgList, nil
}

// Block 当消息同步发生了错误或者用户主动在手机上退出，该方法会立即返回，否则会一直阻塞
func (b *Bot) Block() error {
	if b.self == nil {
		return errors.New("`Block` must be called after user login")
	}
	<-b.Context().Done()
	return b.CrashReason()
}

// Exit 主动退出，让 Block 不在阻塞
func (b *Bot) Exit() {
	b.self = nil
	b.cancel()
	if b.LogoutCallBack != nil {
		b.LogoutCallBack(b)
	}
}

// ExitWith 主动退出并且设置退出原因, 可以通过 `CrashReason` 获取退出原因
func (b *Bot) ExitWith(err error) {
	b.err = err
	b.Exit()
}

// CrashReason 获取当前Bot崩溃的原因
func (b *Bot) CrashReason() error {
	return b.err
}

// DumpHotReloadStorage 写入HotReloadStorage
func (b *Bot) DumpHotReloadStorage() error {
	if b.hotReloadStorage == nil {
		return errors.New("HotReloadStorage can not be nil")
	}
	return b.DumpTo(b.hotReloadStorage)
}

// DumpTo 将热登录需要的数据写入到指定的 io.Writer 中
// 注: 写之前最好先清空之前的数据
func (b *Bot) DumpTo(writer io.Writer) error {
	jar := b.Caller.Client.Jar()
	item := HotReloadStorageItem{
		BaseRequest:  b.Storage.Request,
		Jar:          jar,
		LoginInfo:    b.Storage.LoginInfo,
		WechatDomain: b.Caller.Client.Domain,
		SyncKey:      b.Storage.Response.SyncKey,
		UUID:         b.uuid,
	}
	return json.NewEncoder(writer).Encode(item)
}

// IsHot returns true if is hot login otherwise false
func (b *Bot) IsHot() bool {
	return b.hotReloadStorage != nil
}

// UUID returns current UUID of bot
func (b *Bot) UUID() string {
	return b.uuid
}

// Context returns current context of bot
func (b *Bot) Context() context.Context {
	return b.context
}

// NewBot Bot的构造方法
// 接收外部的 context.Context，用于控制Bot的存活
func NewBot(c context.Context) *Bot {
	caller := DefaultCaller()
	// 默认行为为网页版微信模式
	caller.Client.SetMode(normal)
	ctx, cancel := context.WithCancel(c)
	return &Bot{
		Caller:  caller,
		Storage: &Session{},
		context: ctx,
		cancel:  cancel,
	}
}

func New(ctx context.Context) *Bot {
	return NewBot(ctx)
}

// DefaultBot 默认的Bot的构造方法,
// mode不传入默认为 openwechat.Normal,详情见mode
//
//	bot := openwechat.DefaultBot(openwechat.Desktop)
func DefaultBot(prepares ...BotPreparer) *Bot {
	bot := NewBot(context.Background())
	// 获取二维码回调
	bot.UUIDCallback = PrintlnQrcodeUrl
	// 扫码回调
	bot.ScanCallBack = func(_ CheckLoginResponse) {
		log.Println("扫码成功,请在手机上确认登录")
	}
	// 登录回调
	bot.LoginCallBack = func(_ CheckLoginResponse) {
		log.Println("登录成功")
	}
	// 心跳回调函数
	// 默认的行为打印SyncCheckResponse
	bot.SyncCheckCallback = func(resp SyncCheckResponse) {
		log.Printf("RetCode:%s  Selector:%s", resp.RetCode, resp.Selector)
	}
	for _, prepare := range prepares {
		prepare.Prepare(bot)
	}
	return bot
}

func Default(prepares ...BotPreparer) *Bot {
	return DefaultBot(prepares...)
}

// GetQrcodeUrl 通过uuid获取登录二维码的url
func GetQrcodeUrl(uuid string) string {
	return qrcode + uuid
}

// PrintlnQrcodeUrl 打印登录二维码
func PrintlnQrcodeUrl(uuid string) {
	println("访问下面网址扫描二维码登录")
	qrcodeUrl := GetQrcodeUrl(uuid)
	println(qrcodeUrl)

	// browser open the login url
	_ = open(qrcodeUrl)
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "windows":
		cmd, args = "cmd", []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		// "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
