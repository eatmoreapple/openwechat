package openwechat

import (
	"errors"
)

type errorWrapper struct {
	err error
	msg string
}

func (e errorWrapper) Unwrap() error { return e.err }

func (e errorWrapper) Error() string { return e.msg }

func ErrorWrapper(err error, msg string) error {
	return &errorWrapper{msg: msg, err: err}
}

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
