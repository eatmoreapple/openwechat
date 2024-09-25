package openwechat

import (
	"errors"
)

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

var (
	// ErrForbidden 禁止当前账号登录
	ErrForbidden = errors.New("login forbidden")

	// ErrInvalidStorage define invalid storage error
	ErrInvalidStorage = errors.New("invalid storage")

	// NetworkErr define wechat network error
	NetworkErr = errors.New("wechat network error")

	// ErrNoSuchUserFound define no such user found error
	ErrNoSuchUserFound = errors.New("no such user found")

	// ErrLoginTimeout define login timeout error
	ErrLoginTimeout = errors.New("login timeout")

	// ErrWebWxDataTicketNotFound define webwx_data_ticket not found error
	ErrWebWxDataTicketNotFound = errors.New("webwx_data_ticket not found")

	// ErrUserLogout define user logout error
	ErrUserLogout = errors.New("user logout")
)

// Error impl error interface
func (r Ret) Error() string {
	return r.String()
}
