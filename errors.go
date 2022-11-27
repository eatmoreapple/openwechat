package openwechat

import (
	"errors"
)

var NetworkErr = errors.New("wechat network error")

func IsNetworkError(err error) bool {
	return errors.Is(err, NetworkErr)
}

// IgnoreNetworkError 忽略网络请求的错误
func IgnoreNetworkError(errHandler func(err error)) func(error) {
	return func(err error) {
		if !IsNetworkError(err) {
			errHandler(err)
		}
	}
}

// ErrForbidden 禁止当前账号登录
var ErrForbidden = errors.New("login forbidden")
