package openwechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HttpHook 请求上下文钩子
type HttpHook interface {
	// BeforeRequest 将在请求之前调用
	BeforeRequest(req *http.Request)

	// AfterRequest 将在请求之后调用，无论请求成功与否
	AfterRequest(response *http.Response, err error)
}

// HttpHooks 请求上下文钩子列表
type HttpHooks []HttpHook

// BeforeRequest 将在请求之前调用
func (h HttpHooks) BeforeRequest(req *http.Request) {
	if len(h) == 0 {
		return
	}
	for _, hook := range h {
		hook.BeforeRequest(req)
	}
}

// AfterRequest 将在请求之后调用，无论请求成功与否
func (h HttpHooks) AfterRequest(response *http.Response, err error) {
	if len(h) == 0 {
		return
	}
	for _, hook := range h {
		hook.AfterRequest(response, err)
	}
}

type UserAgentHook struct {
	UserAgent string
}

func (u UserAgentHook) BeforeRequest(req *http.Request) {
	req.Header.Set("User-Agent", u.UserAgent)
}

func (u UserAgentHook) AfterRequest(_ *http.Response, _ error) {}

// defaultUserAgentHook 默认的User-Agent钩子
var defaultUserAgentHook = UserAgentHook{"Mozilla/5.0 (Linux; U; UOS x86_64; zh-cn) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 UOSBrowser/6.0.1.1001"}

// Client http请求客户端
// 客户端需要维持Session会话
type Client struct {
	// 设置一些client的请求行为
	// see normalMode desktopMode
	mode Mode

	// client http客户端
	client *http.Client

	// Domain 微信服务器请求域名
	// 这个参数会在登录成功后被赋值
	// 之后所有的请求都会使用这个域名
	// 在登录热登录和扫码登录时会被重新赋值
	Domain WechatDomain

	// HttpHooks 请求上下文钩子
	HttpHooks HttpHooks

	// MaxRetryTimes 最大重试次数
	MaxRetryTimes int
}

// NewClient 创建一个新的客户端
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		panic("http client is nil")
	}
	client := &Client{client: httpClient}
	client.MaxRetryTimes = 1
	client.SetCookieJar(NewJar())
	return client
}

// DefaultClient 自动存储cookie
// 设置客户端不自动跳转
func DefaultClient() *Client {
	httpClient := &http.Client{
		// 设置客户端不自动跳转
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		// 设置30秒超时
		// 因为微信同步消息时是一个时间长达25秒的长轮训
		Timeout: 30 * time.Second,
	}
	client := NewClient(httpClient)
	client.AddHttpHook(defaultUserAgentHook)
	client.MaxRetryTimes = 5
	return client
}

// AddHttpHook 添加一个请求上下文钩子
func (c *Client) AddHttpHook(hooks ...HttpHook) {
	c.HttpHooks = append(c.HttpHooks, hooks...)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	// 确保请求能够被执行
	if c.MaxRetryTimes <= 0 {
		c.MaxRetryTimes = 1
	}
	var (
		resp        *http.Response
		err         error
		requestBody *bytes.Reader
	)

	c.HttpHooks.BeforeRequest(req)
	defer func() { c.HttpHooks.AfterRequest(resp, err) }()
	if req.Body != nil {
		rawBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("io.ReadAll: %w", err)
		}
		requestBody = bytes.NewReader(rawBody)
	}
	for i := 0; i < c.MaxRetryTimes; i++ {
		if requestBody != nil {
			_, err := requestBody.Seek(0, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("requestBody.Seek: %w", err)
			}
			req.Body = io.NopCloser(requestBody)
		}
		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
	}
	if err != nil {
		err = errors.Join(NetworkErr, err)
	}
	return resp, err
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.do(req)
}

// Jar 返回当前client的 http.CookieJar
// this http.CookieJar must be *Jar type
func (c *Client) Jar() *Jar {
	return c.client.Jar.(*Jar)
}

// SetCookieJar 设置cookieJar
// 这里限制了cookieJar必须是Jar类型
// 否则进行cookie序列化的时候因为字段的私有性无法进行所有字段的导出
func (c *Client) SetCookieJar(jar *Jar) {
	c.client.Jar = jar
}

// HTTPClient 返回http.Client
// 用于自定义http.Client的行为，如设置超时时间、设置代理、设置TLS配置等
func (c *Client) HTTPClient() *http.Client {
	return c.client
}

// GetLoginUUID 获取登录的uuid
func (c *Client) GetLoginUUID(ctx context.Context) (*http.Response, error) {
	req, err := c.mode.BuildGetLoginUUIDRequest(ctx)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// GetLoginQrcode 获取登录的二维吗
func (c *Client) GetLoginQrcode(ctx context.Context, uuid string) (*http.Response, error) {
	path := qrcode + uuid
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

// CheckLogin 检查是否登录
func (c *Client) CheckLogin(ctx context.Context, uuid, tip string) (*http.Response, error) {
	path, err := url.Parse(login)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	params := url.Values{}
	params.Add("r", strconv.FormatInt(now/1579, 10))
	params.Add("_", strconv.FormatInt(now, 10))
	params.Add("loginicon", "true")
	params.Add("uuid", uuid)
	params.Add("tip", tip)
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// GetLoginInfo 请求获取LoginInfo
func (c *Client) GetLoginInfo(ctx context.Context, path *url.URL) (*http.Response, error) {
	req, err := c.mode.BuildGetLoginInfoRequest(ctx, path.String())
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// WebInit 请求获取初始化信息
func (c *Client) WebInit(ctx context.Context, request *BaseRequest) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxinit)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("_", fmt.Sprintf("%d", time.Now().Unix()))
	path.RawQuery = params.Encode()
	content := struct{ BaseRequest *BaseRequest }{BaseRequest: request}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientCommonOptions struct {
	BaseRequest     *BaseRequest
	WebInitResponse *WebInitResponse
	LoginInfo       *LoginInfo
}

type ClientWebWxStatusNotifyOptions ClientCommonOptions

// WebWxStatusNotify 通知手机已登录
func (c *Client) WebWxStatusNotify(ctx context.Context, opt *ClientWebWxStatusNotifyOptions) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxstatusnotify)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	username := opt.WebInitResponse.User.UserName
	content := map[string]interface{}{
		"BaseRequest":  opt.BaseRequest,
		"ClientMsgId":  time.Now().Unix(),
		"Code":         3,
		"FromUserName": username,
		"ToUserName":   username,
	}
	path.RawQuery = params.Encode()
	buffer, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientSyncCheckOptions ClientCommonOptions

// SyncCheck 异步检查是否有新的消息返回
func (c *Client) SyncCheck(ctx context.Context, opt *ClientSyncCheckOptions) (*http.Response, error) {
	path, err := url.Parse(c.Domain.SyncHost() + synccheck)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("r", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	params.Add("skey", opt.LoginInfo.SKey)
	params.Add("sid", opt.LoginInfo.WxSid)
	params.Add("uin", strconv.FormatInt(opt.LoginInfo.WxUin, 10))
	params.Add("deviceid", opt.BaseRequest.DeviceID)
	params.Add("_", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var syncKeyStringSlice = make([]string, opt.WebInitResponse.SyncKey.Count)
	// 将SyncKey里面的元素按照特定的格式拼接起来
	for index, item := range opt.WebInitResponse.SyncKey.List {
		i := fmt.Sprintf("%d_%d", item.Key, item.Val)
		syncKeyStringSlice[index] = i
	}
	syncKey := strings.Join(syncKeyStringSlice, "|")
	params.Add("synckey", syncKey)
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// WebWxGetContact 获取联系人信息
func (c *Client) WebWxGetContact(ctx context.Context, sKey string, reqs int64) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxgetcontact)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("r", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	params.Add("skey", sKey)
	params.Add("seq", strconv.FormatInt(reqs, 10))
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// WebWxBatchGetContact 获取联系人详情
func (c *Client) WebWxBatchGetContact(ctx context.Context, members Members, request *BaseRequest) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxbatchgetcontact)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("type", "ex")
	params.Add("r", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	path.RawQuery = params.Encode()
	list := NewUserDetailItemList(members)
	content := map[string]interface{}{
		"BaseRequest": request,
		"Count":       members.Count(),
		"List":        list,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientWebWxSyncOptions ClientCommonOptions

// WebWxSync 获取消息接口
func (c *Client) WebWxSync(ctx context.Context, opt *ClientWebWxSyncOptions) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxsync)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("sid", opt.LoginInfo.WxSid)
	params.Add("skey", opt.LoginInfo.SKey)
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"BaseRequest": opt.BaseRequest,
		"SyncKey":     opt.WebInitResponse.SyncKey,
		"rr":          strconv.FormatInt(time.Now().Unix(), 10),
	}
	reader, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

// 发送消息
func (c *Client) sendMessage(ctx context.Context, request *BaseRequest, url string, msg *SendMessage) (*http.Response, error) {
	content := map[string]interface{}{
		"BaseRequest": request,
		"Msg":         msg,
		"Scene":       0,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientWebWxSendMsgOptions struct {
	LoginInfo   *LoginInfo
	BaseRequest *BaseRequest
	Message     *SendMessage
}

// WebWxSendMsg 发送文本消息
func (c *Client) WebWxSendMsg(ctx context.Context, opt *ClientWebWxSendMsgOptions) (*http.Response, error) {
	opt.Message.Type = MsgTypeText
	path, err := url.Parse(c.Domain.BaseHost() + webwxsendmsg)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	path.RawQuery = params.Encode()
	return c.sendMessage(ctx, opt.BaseRequest, path.String(), opt.Message)
}

// WebWxSendMsg 发送表情消息
func (c *Client) WebWxSendEmoticon(ctx context.Context, opt *ClientWebWxSendMsgOptions) (*http.Response, error) {
	opt.Message.Type = MsgTypeText
	path, err := url.Parse(c.Domain.BaseHost() + webwxsendemoticon)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "sys")
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	path.RawQuery = params.Encode()
	return c.sendMessage(ctx, opt.BaseRequest, path.String(), opt.Message)
}

// WebWxGetHeadImg 获取用户的头像
func (c *Client) WebWxGetHeadImg(ctx context.Context, user *User) (*http.Response, error) {
	var path string
	if user.HeadImgUrl != "" {
		path = c.Domain.BaseHost() + user.HeadImgUrl
	} else {
		params := url.Values{}
		params.Add("username", user.UserName)
		params.Add("skey", user.self.bot.Storage.Request.Skey)
		params.Add("type", "big")
		params.Add("chatroomid", user.EncryChatRoomId)
		params.Add("seq", "0")
		URL, err := url.Parse(c.Domain.BaseHost() + webwxgeticon)
		if err != nil {
			return nil, err
		}
		URL.RawQuery = params.Encode()
		path = URL.String()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

type webWxCheckUploadRequest struct {
	BaseRequest  *BaseRequest `json:"BaseRequest"`
	FileMd5      string       `json:"FileMd5"`
	FileName     string       `json:"FileName"`
	FileSize     int64        `json:"FileSize"`
	FileType     uint8        `json:"FileType"`
	FromUserName string       `json:"FromUserName"`
	ToUserName   string       `json:"ToUserName"`
}

type webWxCheckUploadResponse struct {
	BaseResponse  BaseResponse `json:"BaseResponse"`
	MediaId       string       `json:"MediaId"`
	AESKey        string       `json:"AESKey"`
	Signature     string       `json:"Signature"`
	EntryFileName string       `json:"EncryFileName"`
}

func (c *Client) webWxCheckUploadRequest(ctx context.Context, req webWxCheckUploadRequest) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxcheckupload)
	if err != nil {
		return nil, err
	}
	body, err := jsonEncode(req)
	if err != nil {
		return nil, err
	}
	reqs, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	reqs.Header.Add("Content-Type", jsonContentType)
	return c.Do(reqs)
}

type uploadMediaRequest struct {
	UploadType    uint8        `json:"UploadType"`
	BaseRequest   *BaseRequest `json:"BaseRequest"`
	ClientMediaId int64        `json:"ClientMediaId"`
	TotalLen      int64        `json:"TotalLen"`
	StartPos      int          `json:"StartPos"`
	DataLen       int64        `json:"DataLen"`
	MediaType     uint8        `json:"MediaType"`
	FromUserName  string       `json:"FromUserName"`
	ToUserName    string       `json:"ToUserName"`
	FileMd5       string       `json:"FileMd5"`
	AESKey        string       `json:"AESKey,omitempty"`
	Signature     string       `json:"Signature,omitempty"`
}

type ClientWebWxUploadMediaByChunkOptions struct {
	FromUserName     string
	ToUserName       string
	BaseRequest      *BaseRequest
	LoginInfo        *LoginInfo
	Filename         string
	FileMD5          string
	FileSize         int64
	LastModifiedDate time.Time
	AESKey           string
	Signature        string
}

type UploadFile interface {
	io.ReaderAt
	io.ReadSeeker
}

// WebWxUploadMediaByChunk 分块上传文件
// TODO 优化掉这个函数
func (c *Client) WebWxUploadMediaByChunk(ctx context.Context, file UploadFile, opt *ClientWebWxUploadMediaByChunkOptions) (*http.Response, error) {
	// 获取文件上传的类型
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	contentType, err := GetFileContentType(file)
	if err != nil {
		return nil, err
	}

	filename := opt.Filename

	if ext := filepath.Ext(filename); ext == "" {
		names := strings.Split(contentType, "/")
		filename = filename + "." + names[len(names)-1]
	}

	// 获取文件的类型
	mediaType := messageType(filename)

	path, err := url.Parse(c.Domain.FileHost() + webwxuploadmedia)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("f", "json")

	path.RawQuery = params.Encode()

	cookies := c.Jar().Cookies(path)

	webWxDataTicket, err := wxDataTicket(cookies)
	if err != nil {
		return nil, err
	}

	uploadMediaRequest := &uploadMediaRequest{
		UploadType:    2,
		BaseRequest:   opt.BaseRequest,
		ClientMediaId: time.Now().Unix() * 1e4,
		TotalLen:      opt.FileSize,
		StartPos:      0,
		DataLen:       opt.FileSize,
		MediaType:     4,
		FromUserName:  opt.FromUserName,
		ToUserName:    opt.ToUserName,
		FileMd5:       opt.FileMD5,
		AESKey:        opt.AESKey,
		Signature:     opt.Signature,
	}

	uploadMediaRequestByte, err := json.Marshal(uploadMediaRequest)
	if err != nil {
		return nil, err
	}

	// 计算上传文件的次数
	chunks := int((opt.FileSize + chunkSize - 1) / chunkSize)

	content := map[string]string{
		"id":                "WU_FILE_0",
		"name":              filename,
		"type":              contentType,
		"lastModifiedDate":  opt.LastModifiedDate.Format(TimeFormat),
		"size":              strconv.FormatInt(opt.FileSize, 10),
		"mediatype":         mediaType,
		"webwx_data_ticket": webWxDataTicket,
		"pass_ticket":       opt.LoginInfo.PassTicket,
	}

	if chunks > 1 {
		content["chunks"] = strconv.Itoa(chunks)
	}

	var (
		resp       *http.Response
		formBuffer = bytes.NewBuffer(nil)
	)

	upload := func(chunk int, fileReader io.Reader) error {
		if chunks > 1 {
			content["chunk"] = strconv.Itoa(chunk)
		}
		formBuffer.Reset()

		writer := multipart.NewWriter(formBuffer)

		// write form data
		{
			if err = writer.WriteField("uploadmediarequest", string(uploadMediaRequestByte)); err != nil {
				return err
			}
			for k, v := range content {
				if err := writer.WriteField(k, v); err != nil {
					return err
				}
			}

			// create form file
			fileWriter, err := writer.CreateFormFile("filename", filename)
			if err != nil {
				return err
			}
			if _, err = io.Copy(fileWriter, fileReader); err != nil {
				return err
			}
			if err = writer.Close(); err != nil {
				return err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), formBuffer)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err = c.Do(req)
		if err != nil {
			return err
		}

		// parse response error
		{
			isLastTime := chunk+1 == chunks
			if !isLastTime {
				defer func() { _ = resp.Body.Close() }()
				parser := MessageResponseParser{Reader: resp.Body}
				err = parser.Err()
			}
		}
		return err
	}

	// 分块上传
	for chunk := 0; chunk < chunks; chunk++ {
		// chunk reader
		selectionReader := io.NewSectionReader(file, int64(chunk)*chunkSize, chunkSize)
		// try to upload
		if err = upload(chunk, selectionReader); err != nil {
			return nil, err
		}
	}
	// 将最后一次携带文件信息的response返回
	return resp, err
}

// WebWxSendMsgImg 发送图片
// 这个接口依赖上传文件的接口
// 发送的图片必须是已经成功上传的图片
func (c *Client) WebWxSendMsgImg(ctx context.Context, opt *ClientWebWxSendMsgOptions) (*http.Response, error) {
	opt.Message.Type = MsgTypeImage
	path, err := url.Parse(c.Domain.BaseHost() + webwxsendmsgimg)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "async")
	params.Add("f", "json")
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	path.RawQuery = params.Encode()
	return c.sendMessage(ctx, opt.BaseRequest, path.String(), opt.Message)
}

// WebWxSendAppMsg 发送文件信息
func (c *Client) WebWxSendAppMsg(ctx context.Context, msg *SendMessage, request *BaseRequest) (*http.Response, error) {
	msg.Type = AppMessage
	path, err := url.Parse(c.Domain.BaseHost() + webwxsendappmsg)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "async")
	params.Add("f", "json")
	params.Add("lang", "zh_CN")
	path.RawQuery = params.Encode()
	return c.sendMessage(ctx, request, path.String(), msg)
}

type ClientWebWxOplogOption struct {
	RemarkName  string
	UserName    string
	BaseRequest *BaseRequest
}

// WebWxOplog 用户重命名接口
func (c *Client) WebWxOplog(ctx context.Context, opt *ClientWebWxOplogOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxoplog)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("lang", "zh_CN")
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"BaseRequest": opt.BaseRequest,
		"CmdId":       2,
		"RemarkName":  opt.RemarkName,
		"UserName":    opt.UserName,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientWebWxVerifyUserOption struct {
	RecommendInfo RecommendInfo
	VerifyContent string
	BaseRequest   *BaseRequest
	LoginInfo     *LoginInfo
}

// WebWxVerifyUser 添加用户为好友接口
func (c *Client) WebWxVerifyUser(ctx context.Context, opt *ClientWebWxVerifyUserOption) (*http.Response, error) {
	loginInfo := opt.LoginInfo
	path, err := url.Parse(c.Domain.BaseHost() + webwxverifyuser)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("r", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", loginInfo.PassTicket)
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"BaseRequest":    opt.BaseRequest,
		"Opcode":         3,
		"SceneList":      [1]int{33},
		"SceneListCount": 1,
		"VerifyContent":  opt.VerifyContent,
		"VerifyUserList": []interface{}{map[string]string{
			"Value":            opt.RecommendInfo.UserName,
			"VerifyUserTicket": opt.RecommendInfo.Ticket,
		}},
		"VerifyUserListSize": 1,
		"skey":               opt.BaseRequest.Skey,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

// WebWxGetMsgImg 获取图片消息的图片响应
func (c *Client) WebWxGetMsgImg(ctx context.Context, msg *Message, info *LoginInfo) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxgetmsgimg)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("MsgID", msg.MsgId)
	params.Add("skey", info.SKey)
	// params.Add("type", "slave")
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// WebWxGetVoice 获取语音消息的语音响应
func (c *Client) WebWxGetVoice(ctx context.Context, msg *Message, info *LoginInfo) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxgetvoice)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("msgid", msg.MsgId)
	params.Add("skey", info.SKey)
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", path.String())
	req.Header.Add("Range", "bytes=0-")
	return c.Do(req)
}

// WebWxGetVideo 获取视频消息的视频响应
func (c *Client) WebWxGetVideo(ctx context.Context, msg *Message, info *LoginInfo) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxgetvideo)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("msgid", msg.MsgId)
	params.Add("skey", info.SKey)
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", path.String())
	req.Header.Add("Range", "bytes=0-")
	return c.Do(req)
}

// WebWxGetMedia 获取文件消息的文件响应
func (c *Client) WebWxGetMedia(ctx context.Context, msg *Message, info *LoginInfo) (*http.Response, error) {
	path, err := url.Parse(c.Domain.FileHost() + webwxgetmedia)
	if err != nil {
		return nil, err
	}
	cookies := c.Jar().Cookies(path)
	webWxDataTicket, err := wxDataTicket(cookies)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("sender", msg.FromUserName)
	params.Add("mediaid", msg.MediaId)
	params.Add("encryfilename", msg.EncryFileName)
	params.Add("fromuser", strconv.FormatInt(info.WxUin, 10))
	params.Add("pass_ticket", info.PassTicket)
	params.Add("webwx_data_ticket", webWxDataTicket)
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", c.Domain.BaseHost()+"/")
	return c.Do(req)
}

// Logout 用户退出
func (c *Client) Logout(ctx context.Context, info *LoginInfo) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxlogout)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("redirect", "1")
	params.Add("type", "1")
	params.Add("skey", info.SKey)
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

type ClientAddMemberIntoChatRoomOption struct {
	Group            string
	GroupLength      int
	InviteMemberList []string
	BaseRequest      *BaseRequest
	LoginInfo        *LoginInfo
}

// AddMemberIntoChatRoom 添加用户进群聊
func (c *Client) AddMemberIntoChatRoom(ctx context.Context, opt *ClientAddMemberIntoChatRoomOption) (*http.Response, error) {
	if opt.GroupLength >= 40 {
		return c.InviteMemberIntoChatRoom(ctx, opt)
	}
	return c.addMemberIntoChatRoom(ctx, opt)
}

// addMemberIntoChatRoom 添加用户进群聊
func (c *Client) addMemberIntoChatRoom(ctx context.Context, opt *ClientAddMemberIntoChatRoomOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxupdatechatroom)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "addmember")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	params.Add("lang", "zh_CN")
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"ChatRoomName":  opt.Group,
		"BaseRequest":   opt.BaseRequest,
		"AddMemberList": strings.Join(opt.InviteMemberList, ","),
	}
	buffer, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), buffer)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", jsonContentType)
	return c.Do(httpReq)
}

// InviteMemberIntoChatRoom 邀请用户进群聊
func (c *Client) InviteMemberIntoChatRoom(ctx context.Context, opt *ClientAddMemberIntoChatRoomOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxupdatechatroom)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "invitemember")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	params.Add("lang", "zh_CN")
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"ChatRoomName":     opt.Group,
		"BaseRequest":      opt.BaseRequest,
		"InviteMemberList": strings.Join(opt.InviteMemberList, ","),
	}
	buffer, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), buffer)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", jsonContentType)
	return c.Do(httpReq)
}

type ClientRemoveMemberFromChatRoomOption struct {
	Group         string
	DelMemberList []string
	BaseRequest   *BaseRequest
	LoginInfo     *LoginInfo
}

// RemoveMemberFromChatRoom 从群聊中移除用户
func (c *Client) RemoveMemberFromChatRoom(ctx context.Context, opt *ClientRemoveMemberFromChatRoomOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxupdatechatroom)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "delmember")
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	content := map[string]interface{}{
		"ChatRoomName":  opt.Group,
		"BaseRequest":   opt.BaseRequest,
		"DelMemberList": strings.Join(opt.DelMemberList, ","),
	}
	buffer, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), buffer)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", jsonContentType)
	return c.Do(httpReq)
}

// WebWxRevokeMsg 撤回消息
func (c *Client) WebWxRevokeMsg(ctx context.Context, msg *SentMessage, request *BaseRequest) (*http.Response, error) {
	content := map[string]interface{}{
		"BaseRequest": request,
		"ClientMsgId": msg.ClientMsgId,
		"SvrMsgId":    msg.MsgId,
		"ToUserName":  msg.ToUserName,
	}
	buffer, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Domain.BaseHost()+webwxrevokemsg, buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", jsonContentType)
	return c.Do(req)
}

// 校验上传文件
// nolint:unused
func (c *Client) webWxCheckUpload(stat os.FileInfo, request *BaseRequest, fileMd5, fromUserName, toUserName string) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxcheckupload)
	if err != nil {
		return nil, err
	}
	content := map[string]interface{}{
		"BaseRequest":  request,
		"FileMd5":      fileMd5,
		"FileName":     stat.Name(),
		"FileSize":     stat.Size(),
		"FileType":     7,
		"FromUserName": fromUserName,
		"ToUserName":   toUserName,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientWebWxStatusAsReadOption struct {
	LoginInfo *LoginInfo
	Request   *BaseRequest
	Message   *Message
}

func (c *Client) WebWxStatusAsRead(ctx context.Context, opt *ClientWebWxStatusAsReadOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxstatusnotify)
	if err != nil {
		return nil, err
	}
	content := map[string]interface{}{
		"BaseRequest":  opt.Request,
		"DeviceID":     opt.Request.DeviceID,
		"Sid":          opt.Request.Sid,
		"Skey":         opt.Request.Skey,
		"Uin":          opt.LoginInfo.WxUin,
		"ClientMsgId":  time.Now().Unix(),
		"Code":         1,
		"FromUserName": opt.Message.ToUserName,
		"ToUserName":   opt.Message.FromUserName,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientWebWxRelationPinOption struct {
	Request    *BaseRequest
	Op         uint8
	RemarkName string
	UserName   string
}

// WebWxRelationPin 联系人置顶接口
func (c *Client) WebWxRelationPin(ctx context.Context, opt *ClientWebWxRelationPinOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxoplog)
	if err != nil {
		return nil, err
	}
	content := map[string]interface{}{
		"BaseRequest": opt.Request,
		"CmdId":       3,
		"OP":          opt.Op,
		"RemarkName":  opt.RemarkName,
		"UserName":    opt.UserName,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

// WebWxPushLogin 免扫码登陆接口
func (c *Client) WebWxPushLogin(ctx context.Context, uin int64) (*http.Response, error) {
	req, err := c.mode.BuildPushLoginRequest(ctx, c.Domain.BaseHost(), uin)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// WebWxSendVideoMsg 发送视频消息接口
func (c *Client) WebWxSendVideoMsg(ctx context.Context, request *BaseRequest, msg *SendMessage) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxsendvideomsg)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "async")
	params.Add("f", "json")
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", "pass_ticket")
	path.RawQuery = params.Encode()
	return c.sendMessage(ctx, request, path.String(), msg)
}

type ClientWebWxCreateChatRoomOption struct {
	Request   *BaseRequest
	LoginInfo *LoginInfo
	Topic     string
	Friends   []string
}

// WebWxCreateChatRoom 创建群聊
func (c *Client) WebWxCreateChatRoom(ctx context.Context, opt *ClientWebWxCreateChatRoomOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxcreatechatroom)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	params.Add("r", fmt.Sprintf("%d", time.Now().Unix()))
	path.RawQuery = params.Encode()
	count := len(opt.Friends)
	memberList := make([]struct{ UserName string }, count)
	for index, member := range opt.Friends {
		memberList[index] = struct{ UserName string }{member}
	}
	content := map[string]interface{}{
		"BaseRequest": opt.Request,
		"MemberCount": count,
		"MemberList":  memberList,
		"Topic":       opt.Topic,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

type ClientWebWxRenameChatRoomOption struct {
	Request   *BaseRequest
	LoginInfo *LoginInfo
	NewTopic  string
	Group     string
}

// WebWxRenameChatRoom 群组重命名接口
func (c *Client) WebWxRenameChatRoom(ctx context.Context, opt *ClientWebWxRenameChatRoomOption) (*http.Response, error) {
	path, err := url.Parse(c.Domain.BaseHost() + webwxupdatechatroom)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fun", "modtopic")
	params.Add("pass_ticket", opt.LoginInfo.PassTicket)
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"BaseRequest":  opt.Request,
		"ChatRoomName": opt.Group,
		"NewTopic":     opt.NewTopic,
	}
	body, err := jsonEncode(content)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}
