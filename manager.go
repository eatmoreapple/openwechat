package openwechat
//
//import (
//	"encoding/json"
//	"encoding/xml"
//	"errors"
//	"fmt"
//	"io/ioutil"
//	"log"
//)
//
//var DefaultMessageMaxLength uint64 = 200
//
//const (
//	statusSuccess = "200"
//	statusScanned = "201"
//	statusTimeout = "400"
//	statusWait    = "408"
//)
//
//type Bot struct {
//	Caller         *Caller
//	Self           *Self
//	ScanCallback   func(body []byte)
//	LoginCallback  func(body []byte)
//	storage        WechatStorage
//	messageHandler MessageHandler
//	notAlive       bool
//	err            error
//}
//
//func (m *Bot) GetLoginUUID() (uuid string, err error) {
//	return m.Caller.GetLoginUUID()
//}
//
//func (m *Bot) checkLogin(uuid string) (body []byte, err error) {
//	resp, err := m.Client.CheckLogin(uuid)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//	data, err := ioutil.ReadAll(resp.Body)
//	if
//
//	err != nil {
//		return nil, err
//	}
//	return data, nil
//}
//
//func (m *Bot) CheckLogin(uuid string) error {
//	var (
//		body []byte
//		err  error
//	)
//	for {
//		body, err = m.checkLogin(uuid)
//		if err != nil {
//			return err
//		}
//		results := statusCodeRegexp.FindSubmatch(body)
//		if len(results) != 2 {
//			return errors.New("login status code does not match")
//		}
//		code := string(results[1])
//		switch code {
//		case statusSuccess:
//			return m.loginCallback(body)
//		case statusScanned:
//			if m.ScanCallback != nil {
//				m.ScanCallback(body)
//			}
//		case statusWait:
//			log.Println(string(body))
//		case statusTimeout:
//			return errors.New("login time out")
//		default:
//			return errors.New("unknow code found " + code)
//		}
//	}
//}
//
//func (m *Bot) getLoginInfo(body []byte) error {
//	resp, err := m.Client.GetLoginInfo(body)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//	data, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return err
//	}
//	var loginInfo LoginInfo
//	if err = xml.Unmarshal(data, &loginInfo); err != nil {
//		return err
//	}
//	if loginInfo.Ret != 0 {
//		return errors.New(loginInfo.Message)
//	}
//	m.storage.SetLoginInfo(loginInfo)
//	return nil
//}
//
//func (m *Bot) webInit() error {
//	loginInfo := m.storage.GetLoginInfo()
//	baseRequest := BaseRequest{
//		Uin:      loginInfo.WxUin,
//		Sid:      loginInfo.WxSid,
//		Skey:     loginInfo.SKey,
//		DeviceID: GetRandomDeviceId(),
//	}
//	m.storage.SetBaseRequest(baseRequest)
//	resp, err := m.Client.WebInit(m.storage)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//	data, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return err
//	}
//	var webInitResponse WebInitResponse
//	if err = json.Unmarshal(data, &webInitResponse); err != nil {
//		return err
//	}
//	m.storage.SetWebInitResponse(webInitResponse)
//	m.Self = &Self{User: &webInitResponse.User, Manager: m}
//	return nil
//}
//
//func (m *Bot) WebWxStatusNotify() error {
//	resp, err := m.Client.WebWxStatusNotify(m.storage)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//	data, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return err
//	}
//	var item map[string]interface{}
//	err = json.Unmarshal(data, &item)
//	if err != nil {
//		return err
//	}
//	if request, ok := item["BaseResponse"].(map[string]interface{}); ok {
//		if ret, exist := request["Ret"]; exist {
//			if ret, ok := ret.(float64); ok {
//				if ret == 0 {
//					return nil
//				}
//			}
//		}
//	}
//	return errors.New("web status notify failed")
//}
//
//func (m *Bot) SyncCheck() error {
//	for m.Alive() {
//		resp, err := m.Client.SyncCheck(m.storage)
//		if err != nil {
//			return err
//		}
//		data, err := ioutil.ReadAll(resp.Body)
//		fmt.Println(string(data))
//		resp.Body.Close()
//		if err != nil {
//			return err
//		}
//		results := syncCheckRegexp.FindSubmatch(data)
//		if len(results) != 3 {
//			return errors.New("parse sync key failed")
//		}
//		code, _ := results[1], results[2]
//		switch string(code) {
//		case "0":
//			if err = m.getMessage(); err != nil {
//				return err
//			}
//		case "1101":
//			return errors.New("logout")
//		}
//		return fmt.Errorf("error ret code: %s", string(code))
//	}
//	return nil
//}
//
//func (m *Bot) getMessage() error {
//	resp, err := m.Client.GetMessage(m.storage)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//	data, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return err
//	}
//	var syncKey struct{ SyncKey SyncKey }
//	if err = json.Unmarshal(data, &syncKey); err != nil {
//		return err
//	}
//	webInitResponse := m.storage.GetWebInitResponse()
//	webInitResponse.SyncKey = syncKey.SyncKey
//	m.storage.SetWebInitResponse(webInitResponse)
//	var messageList MessageList
//	if err = json.Unmarshal(data, &messageList); err != nil {
//		return err
//	}
//	for _, message := range messageList.AddMsgList {
//		message.ClientManager = m
//		m.messageHandler.ReceiveMessage(message)
//	}
//	return nil
//}
//
//func (m *Bot) loginCallback(body []byte) error {
//	var err error
//	if m.LoginCallback != nil {
//		m.LoginCallback(body)
//	}
//	if err = m.getLoginInfo(body); err != nil {
//		return err
//	}
//	if err = m.webInit(); err != nil {
//		return err
//	}
//	if err = m.WebWxStatusNotify(); err != nil {
//		return err
//	}
//	go func() {
//		if err := m.SyncCheck(); err != nil {
//			m.exit(err)
//		}
//	}()
//	return err
//}
//
//func (m *Bot) Alive() bool {
//	return !m.notAlive
//}
//
//func (m *Bot) Err() error {
//	return m.err
//}
//
//func (m *Bot) exit(err error) {
//	m.notAlive = true
//	m.err = err
//}
