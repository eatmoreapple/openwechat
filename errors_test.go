package openwechat

import (
	"errors"
	"testing"
)

func TestIsNetworkError(t *testing.T) {
	var err = errors.New("test error")
	err = ErrorWrapper(NetworkErr, err.Error())
	if !IsNetworkError(err) {
		t.Error("err is not network error")
	}
}
