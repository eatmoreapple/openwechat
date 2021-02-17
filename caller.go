package openwechat

import (
	"errors"
	"fmt"
	"os"
)

type Caller struct {
	Client *Client
}

func NewCaller(client *Client) *Caller {
	return &Caller{Client: client}
}

func DefaultCaller() *Caller {
	return NewCaller(DefaultClient())
}

func (c *Caller) GetLoginUUID() (string, error) {
	resp := NewReturnResponse(c.Client.GetLoginUUID())
	if resp.Err() != nil {
		return "", resp.Err()
	}
	defer resp.Body.Close()
	data, err := resp.ReadAll()
	if err != nil {
		return "", err
	}
	results := uuidRegexp.FindSubmatch(data)
	if len(results) != 2 {
		return "", errors.New("uuid does not match")
	}
	return string(results[1]), nil
}

func (c *Caller) CheckLogin(uuid string) (*CheckLoginResponse, error) {
	resp := NewReturnResponse(c.Client.CheckLogin(uuid))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	defer resp.Body.Close()
	data, err := resp.ReadAll()
	if err != nil {
		return nil, err
	}
	results := statusCodeRegexp.FindSubmatch(data)
	if len(results) != 2 {
		return nil, nil
	}
	code := string(results[1])
	return &CheckLoginResponse{Code: code, Raw: data}, nil
}

func (c *Caller) GetLoginInfo(body []byte) (*LoginInfo, error) {
	results := redirectUriRegexp.FindSubmatch(body)
	if len(results) != 2 {
		return nil, errors.New("redirect url does not match")
	}
	path := string(results[1])
	resp := NewReturnResponse(c.Client.GetLoginInfo(path))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	defer resp.Body.Close()
	var loginInfo LoginInfo
	if err := resp.ScanXML(&loginInfo); err != nil {
		return nil, err
	}
	return &loginInfo, nil
}

func (c *Caller) WebInit(request BaseRequest) (*WebInitResponse, error) {
	resp := NewReturnResponse(c.Client.WebInit(request))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	var webInitResponse WebInitResponse
	defer resp.Body.Close()
	if err := resp.ScanJSON(&webInitResponse); err != nil {
		return nil, err
	}
	return &webInitResponse, nil
}

func (c *Caller) WebWxStatusNotify(request BaseRequest, response WebInitResponse, info LoginInfo) error {
	resp := NewReturnResponse(c.Client.WebWxStatusNotify(request, response, info))
	if resp.Err() != nil {
		return resp.Err()
	}
	var item struct{ BaseResponse BaseResponse }
	defer resp.Body.Close()
	if err := resp.ScanJSON(&item); err != nil {
		return err
	}
	if !item.BaseResponse.Ok() {
		return item.BaseResponse
	}
	return nil
}

func (c *Caller) SyncCheck(info LoginInfo, response WebInitResponse) (*SyncCheckResponse, error) {
	resp := NewReturnResponse(c.Client.SyncCheck(info, response))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	defer resp.Body.Close()
	data, err := resp.ReadAll()
	fmt.Println(string(data))
	if err != nil {
		return nil, err
	}
	results := syncCheckRegexp.FindSubmatch(data)
	if len(results) != 3 {
		return nil, errors.New("parse sync key failed")
	}
	retCode, selector := string(results[1]), string(results[2])
	syncCheckResponse := &SyncCheckResponse{RetCode: retCode, Selector: selector}
	return syncCheckResponse, nil
}

func (c *Caller) WebWxGetContact(info LoginInfo) (Members, error) {
	resp := NewReturnResponse(c.Client.WebWxGetContact(info))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	defer resp.Body.Close()
	var item WebWxContactResponse
	if err := resp.ScanJSON(&item); err != nil {
		return nil, err
	}
	if !item.BaseResponse.Ok() {
		return nil, item.BaseResponse
	}
	return item.MemberList, nil
}

func (c *Caller) WebWxBatchGetContact(members Members, request BaseRequest) (Members, error) {
	resp := NewReturnResponse(c.Client.WebWxBatchGetContact(members, request))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	defer resp.Body.Close()
	var item WebWxBatchContactResponse
	if err := resp.ScanJSON(&item); err != nil {
		return nil, err
	}
	if !item.BaseResponse.Ok() {
		return nil, item.BaseResponse
	}
	return item.ContactList, nil
}

func (c *Caller) WebWxSync(request BaseRequest, response WebInitResponse, info LoginInfo) (*WebWxSyncResponse, error) {
	resp := NewReturnResponse(c.Client.WebWxSync(request, response, info))
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	defer resp.Body.Close()
	var webWxSyncResponse WebWxSyncResponse
	if err := resp.ScanJSON(&webWxSyncResponse); err != nil {
		return nil, err
	}
	return &webWxSyncResponse, nil
}

func (c *Caller) WebWxSendMsg(msg *SendMessage, info LoginInfo, request BaseRequest) error {
	resp := NewReturnResponse(c.Client.WebWxSendMsg(msg, info, request))
	return parseBaseResponseError(resp)
}

func (c *Caller) WebWxOplog(request BaseRequest, remarkName, toUserName string) error {
	resp := NewReturnResponse(c.Client.WebWxOplog(request, remarkName, toUserName))
	return parseBaseResponseError(resp)
}

func (c *Caller) WebWxSendImageMsg(file *os.File, request BaseRequest, info LoginInfo, fromUserName, toUserName string) error {
	resp := NewReturnResponse(c.Client.WebWxUploadMedia(file, request, info, fromUserName, toUserName, "image/jpeg", "pic"))
	if resp.Err() != nil {
		return resp.Err()
	}
	defer resp.Body.Close()
	var item struct {
		BaseResponse BaseResponse
		MediaId      string
	}
	if err := resp.ScanJSON(&item); err != nil {
		return err
	}
	if !item.BaseResponse.Ok() {
		return item.BaseResponse
	}
	msg := NewMediaSendMessage(ImageMessage, fromUserName, toUserName, item.MediaId)
	resp = NewReturnResponse(c.Client.WebWxSendMsgImg(msg, request, info))
	return parseBaseResponseError(resp)
}

func parseBaseResponseError(resp *ReturnResponse) error {
	if resp.Err() != nil {
		return resp.Err()
	}
	defer resp.Body.Close()
	var item struct{ BaseResponse BaseResponse }
	if err := resp.ScanJSON(&item); err != nil {
		return err
	}
	if !item.BaseResponse.Ok() {
		return item.BaseResponse
	}
	return nil
}
