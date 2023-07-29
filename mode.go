package openwechat

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Mode interface {
	GetLoginUUID(ctx context.Context, client *Client) (*http.Response, error)
	GetLoginInfo(ctx context.Context, client *Client, path string) (*http.Response, error)
	PushLogin(ctx context.Context, client *Client, uin int64) (*http.Response, error)
}

var (
	// normal 网页版模式
	normal Mode = normalMode{}

	// desktop 桌面模式，uos electron套壳
	desktop Mode = desktopMode{}
)

type normalMode struct{}

func (n normalMode) PushLogin(ctx context.Context, client *Client, uin int64) (*http.Response, error) {
	path, err := url.Parse(client.Domain.BaseHost() + webwxpushloginurl)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("uin", strconv.FormatInt(uin, 10))
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func (n normalMode) GetLoginUUID(ctx context.Context, client *Client) (*http.Response, error) {
	path, err := url.Parse(jslogin)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	redirectUrl, err := url.Parse(webwxnewloginpage)
	if err != nil {
		return nil, err
	}
	params.Add("redirect_uri", redirectUrl.String())
	params.Add("appid", appId)
	params.Add("fun", "new")
	params.Add("lang", "zh_CN")
	params.Add("_", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	path.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func (n normalMode) GetLoginInfo(ctx context.Context, client *Client, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

type desktopMode struct{}

func (n desktopMode) GetLoginUUID(ctx context.Context, client *Client) (*http.Response, error) {
	path, err := url.Parse(jslogin)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	redirectUrl, err := url.Parse(webwxnewloginpage)
	if err != nil {
		return nil, err
	}
	p := url.Values{"mod": {"desktop"}}
	redirectUrl.RawQuery = p.Encode()
	params.Add("redirect_uri", redirectUrl.String())
	params.Add("appid", appId)
	params.Add("fun", "new")
	params.Add("lang", "zh_CN")
	params.Add("_", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	path.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func (n desktopMode) GetLoginInfo(ctx context.Context, client *Client, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("client-version", uosPatchClientVersion)
	req.Header.Add("extspam", uosPatchExtspam)
	return client.Do(req)
}

func (n desktopMode) PushLogin(ctx context.Context, client *Client, uin int64) (*http.Response, error) {
	path, err := url.Parse(client.Domain.BaseHost() + webwxpushloginurl)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("uin", strconv.FormatInt(uin, 10))
	params.Add("mod", "desktop")
	path.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}
