package openwechat

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	*http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{Client: client}
}

func DefaultClient() *Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}
	return NewClient(client)
}

func (c *Client) GetLoginUUID() (*http.Response, error) {
	path, _ := url.Parse(jsLoginUrl)
	params := url.Values{}
	params.Add("appid", appId)
	params.Add("redirect_uri", webWxNewLoginPage)
	params.Add("fun", "new")
	params.Add("lang", "zh_CN")
	params.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
	path.RawQuery = params.Encode()
	return c.Get(path.String())
}

func (c *Client) GetLoginQrcode(uuid string) (*http.Response, error) {
	path := qrcodeUrl + uuid
	return c.Get(path)
}

func (c *Client) CheckLogin(uuid string) (*http.Response, error) {
	path, _ := url.Parse(loginUrl)
	now := time.Now().Unix()
	params := url.Values{}
	params.Add("r", strconv.FormatInt(now/1579, 10))
	params.Add("_", strconv.FormatInt(now, 10))
	params.Add("loginicon", "true")
	params.Add("uuid", uuid)
	params.Add("tip", "0")
	path.RawQuery = params.Encode()
	return c.Get(path.String())
}

func (c *Client) GetLoginInfo(path string) (*http.Response, error) {
	return c.Get(path)
}

func (c *Client) WebInit(request BaseRequest) (*http.Response, error) {
	path, _ := url.Parse(webWxInitUrl)
	params := url.Values{}
	params.Add("_", fmt.Sprintf("%d", time.Now().Unix()))
	path.RawQuery = params.Encode()
	content := struct{ BaseRequest BaseRequest }{BaseRequest: request}
	body, err := ToBuffer(content)
	if err != nil {
		return nil, err
	}
	return c.Post(path.String(), jsonContentType, body)
}

func (c *Client) WebWxStatusNotify(request BaseRequest, response WebInitResponse, info LoginInfo) (*http.Response, error) {
	path, _ := url.Parse(webWxStatusNotifyUrl)
	params := url.Values{}
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", info.PassTicket)
	username := response.User.UserName
	content := map[string]interface{}{
		"BaseRequest":  request,
		"ClientMsgId":  time.Now().Unix(),
		"Code":         3,
		"FromUserName": username,
		"ToUserName":   username,
	}
	path.RawQuery = params.Encode()
	buffer, _ := ToBuffer(content)
	req, _ := http.NewRequest(http.MethodPost, path.String(), buffer)
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

func (c *Client) SyncCheck(info LoginInfo, response WebInitResponse) (*http.Response, error) {
	path, _ := url.Parse(syncCheckUrl)
	params := url.Values{}
	params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
	params.Add("skey", info.SKey)
	params.Add("sid", info.WxSid)
	params.Add("uin", strconv.Itoa(info.WxUin))
	params.Add("deviceid", GetRandomDeviceId())
	params.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
	syncKeyStringSlice := make([]string, 0)
	for _, item := range response.SyncKey.List {
		i := fmt.Sprintf("%d_%d", item.Key, item.Val)
		syncKeyStringSlice = append(syncKeyStringSlice, i)
	}
	syncKey := strings.Join(syncKeyStringSlice, "|")
	params.Add("synckey", syncKey)
	path.RawQuery = params.Encode()
	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
	req.Header.Add("User-Agent", "Mozilla/5.0")
	return c.Do(req)
}

func (c *Client) WebWxGetContact(info LoginInfo) (*http.Response, error) {
	path, _ := url.Parse(webWxGetContactUrl)
	params := url.Values{}
	params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
	params.Add("skey", info.SKey)
	params.Add("req", "0")
	path.RawQuery = params.Encode()
	return c.Get(path.String())
}

func (c *Client) WebWxBatchGetContact(members Members, request BaseRequest) (*http.Response, error) {
	path, _ := url.Parse(webWxBatchGetContactUrl)
	params := url.Values{}
	params.Add("type", "ex")
	params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
	path.RawQuery = params.Encode()
	list := NewUserDetailItemList(members)
	content := map[string]interface{}{
		"BaseRequest": request,
		"Count":       members.Count(),
		"List":        list,
	}
	body, _ := ToBuffer(content)
	req, _ := http.NewRequest(http.MethodPost, path.String(), body)
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

func (c *Client) WebWxSync(request BaseRequest, response WebInitResponse, info LoginInfo) (*http.Response, error) {
	path, _ := url.Parse(webWxSyncUrl)
	params := url.Values{}
	params.Add("sid", info.WxSid)
	params.Add("skey", info.SKey)
	params.Add("pass_ticket", info.PassTicket)
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"BaseRequest": request,
		"SyncKey":     response.SyncKey,
		"rr":          strconv.FormatInt(time.Now().Unix(), 10),
	}
	data, _ := json.Marshal(content)
	body := bytes.NewBuffer(data)
	req, _ := http.NewRequest(http.MethodPost, path.String(), body)
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

func (c *Client) sendMessage(request BaseRequest, url string, msg *SendMessage) (*http.Response, error) {
	content := map[string]interface{}{
		"BaseRequest": request,
		"Msg":         msg,
		"Scene":       0,
	}
	body, _ := ToBuffer(content)
	req, _ := http.NewRequest(http.MethodPost, url, body)
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

func (c *Client) WebWxSendMsg(msg *SendMessage, info LoginInfo, request BaseRequest) (*http.Response, error) {
	msg.Type = TextMessage
	path, _ := url.Parse(webWxSendMsgUrl)
	params := url.Values{}
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", info.PassTicket)
	path.RawQuery = params.Encode()
	return c.sendMessage(request, path.String(), msg)
}

func (c *Client) WebWxGetHeadImg(headImageUrl string) (*http.Response, error) {
	path := baseUrl + headImageUrl
	return c.Get(path)
}

func (c *Client) WebWxUploadMedia(file *os.File, request BaseRequest, info LoginInfo, forUserName, toUserName, contentType, mediaType string) (*http.Response, error) {
	path, _ := url.Parse(webWxUpLoadMediaUrl)
	params := url.Values{}
	params.Add("f", "json")
	path.RawQuery = params.Encode()
	sate, err := file.Stat()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fileMd5 := fmt.Sprintf("%x", md5.Sum(data))
	cookies := c.Jar.Cookies(path)
	uploadMediaRequest := map[string]interface{}{
		"UploadType":    2,
		"BaseRequest":   request,
		"ClientMediaId": time.Now().Unix() * 1e4,
		"TotalLen":      sate.Size(),
		"StartPos":      0,
		"DataLen":       sate.Size(),
		"MediaType":     4,
		"FromUserName":  forUserName,
		"ToUserName":    toUserName,
		"FileMd5":       fileMd5,
	}
	uploadMediaRequestByte, err := json.Marshal(uploadMediaRequest)
	if err != nil {
		return nil, err
	}
	content := map[string]interface{}{
		"id":                "WU_FILE_0",
		"name":              file.Name(),
		"type":              contentType,
		"lastModifiedDate":  time.Now().Format(http.TimeFormat),
		"size":              sate.Size(),
		"mediatype":         mediaType,
		"webwx_data_ticket": getWebWxDataTicket(cookies),
		"pass_ticket":       info.PassTicket,
	}
	body, err := ToBuffer(content)
	if err != nil {
		return nil, err
	}
	writer := multipart.NewWriter(body)
	if err = writer.WriteField("uploadmediarequest", string(uploadMediaRequestByte)); err != nil {
		return nil, err
	}
	if w, err := writer.CreateFormFile("filename", file.Name()); err != nil {
		return nil, err
	} else {
		if _, err = w.Write(data); err != nil {
			return nil, err
		}
	}
	ct := writer.FormDataContentType()
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return c.Post(path.String(), ct, body)
}

func (c *Client) WebWxSendMsgImg(msg *SendMessage, request BaseRequest, info LoginInfo) (*http.Response, error) {
	msg.Type = ImageMessage
	path, _ := url.Parse(webWxSendMsgImgUrl)
	params := url.Values{}
	params.Add("fun", "async")
	params.Add("f", "json")
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", info.PassTicket)
	path.RawQuery = params.Encode()
	return c.sendMessage(request, path.String(), msg)
}

func (c *Client) WebWxSendAppMsg(msg *SendMessage, request BaseRequest) (*http.Response, error) {
	msg.Type = AppMessage
	path, _ := url.Parse(webWxSendAppMsgUrl)
	params := url.Values{}
	params.Add("fun", "async")
	params.Add("f", "json")
	path.RawQuery = params.Encode()
	return c.sendMessage(request, path.String(), msg)
}

func (c *Client) WebWxOplog(request BaseRequest, remarkName, userName string, ) (*http.Response, error) {
	path, _ := url.Parse(webWxOplogUrl)
	params := url.Values{}
	params.Add("lang", "zh_CN")
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"BaseRequest": request,
		"CmdId":       2,
		"RemarkName":  remarkName,
		"UserName":    userName,
	}
	body, _ := ToBuffer(content)
	req, _ := http.NewRequest(http.MethodPost, path.String(), body)
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}

func (c *Client) WebWxVerifyUser(storage WechatStorage, info RecommendInfo, verifyContent string) (*http.Response, error) {
	loginInfo := storage.GetLoginInfo()
	path, _ := url.Parse(webWxVerifyUserUrl)
	params := url.Values{}
	params.Add("r", strconv.FormatInt(time.Now().Unix(), 10))
	params.Add("lang", "zh_CN")
	params.Add("pass_ticket", loginInfo.PassTicket)
	path.RawQuery = params.Encode()
	content := map[string]interface{}{
		"BaseRequest":    storage.GetBaseRequest(),
		"Opcode":         3,
		"SceneList":      []int{33},
		"SceneListCount": 1,
		"VerifyContent":  verifyContent,
		"VerifyUserList": []interface{}{map[string]string{
			"Value":            info.UserName,
			"VerifyUserTicket": info.Ticket,
		}},
		"VerifyUserListSize": 1,
		"skey":               loginInfo.SKey,
	}
	body, _ := ToBuffer(content)
	req, _ := http.NewRequest(http.MethodPost, path.String(), body)
	req.Header.Add("Content-Type", jsonContentType)
	return c.Do(req)
}
