package openwechat

import (
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Mode interface {
	GetLoginUUID(client *Client) (*http.Response, error)
	GetLoginInfo(client *Client, path string) (*http.Response, error)
}

var (
	normal Mode = normalMode{}

	desktop Mode = desktopMode{}
)

type normalMode struct{}

func (n normalMode) GetLoginUUID(client *Client) (*http.Response, error) {
	path, _ := url.Parse(jslogin)
	params := url.Values{}
	redirectUrl, _ := url.Parse(webwxnewloginpage)
	params.Add("redirect_uri", redirectUrl.String())
	params.Add("appid", appId)
	params.Add("fun", "new")
	params.Add("lang", "zh_CN")
	params.Add("_", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))

	path.RawQuery = params.Encode()
	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
	return client.Do(req)
}

func (n normalMode) GetLoginInfo(client *Client, path string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	return client.Do(req)
}

type desktopMode struct{}

func (n desktopMode) GetLoginUUID(client *Client) (*http.Response, error) {
	path, _ := url.Parse(jslogin)
	params := url.Values{}
	redirectUrl, _ := url.Parse(webwxnewloginpage)
	p := url.Values{"mod": {"desktop"}}
	redirectUrl.RawQuery = p.Encode()
	params.Add("redirect_uri", redirectUrl.String())
	params.Add("appid", appId)
	params.Add("fun", "new")
	params.Add("lang", "zh_CN")
	params.Add("_", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))

	path.RawQuery = params.Encode()
	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)
	return client.Do(req)
}

func (n desktopMode) GetLoginInfo(client *Client, path string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Add("client-version", uosPatchClientVersion)
	req.Header.Add("extspam", uosPatchExtspam)
	return client.Do(req)
}
