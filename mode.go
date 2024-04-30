package openwechat

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Mode interface {
	BuildGetLoginUUIDRequest(ctx context.Context) (*http.Request, error)
	BuildGetLoginInfoRequest(ctx context.Context, path string) (*http.Request, error)
	BuildPushLoginRequest(ctx context.Context, host string, uin int64) (*http.Request, error)
}

var (
	// normal 网页版模式
	normal Mode = normalMode{}

	// desktop 桌面模式，uos electron套壳
	desktop Mode = desktopMode{}
)

type normalMode struct{}

func (n normalMode) BuildGetLoginUUIDRequest(ctx context.Context) (*http.Request, error) {
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
	return http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
}

func (n normalMode) BuildGetLoginInfoRequest(ctx context.Context, path string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
}

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

func (n normalMode) BuildPushLoginRequest(ctx context.Context, host string, uin int64) (*http.Request, error) {
	path, err := url.Parse(host + webwxpushloginurl)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("uin", strconv.FormatInt(uin, 10))
	path.RawQuery = params.Encode()
	return http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
}

type desktopMode struct{}

func (n desktopMode) BuildGetLoginUUIDRequest(ctx context.Context) (*http.Request, error) {
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
	return http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
}

func (n desktopMode) BuildGetLoginInfoRequest(ctx context.Context, path string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("client-version", uosPatchClientVersion)
	req.Header.Add("extspam", uosPatchExtspam)
	return req, nil
}

func (n desktopMode) BuildPushLoginRequest(ctx context.Context, host string, uin int64) (*http.Request, error) {
	path, err := url.Parse(host + webwxpushloginurl)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("uin", strconv.FormatInt(uin, 10))
	params.Add("mod", "desktop")
	path.RawQuery = params.Encode()
	return http.NewRequestWithContext(ctx, http.MethodGet, path.String(), nil)
}
