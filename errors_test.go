package openwechat

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNetworkError(t *testing.T) {
	var err = errors.New("test error")
	err = fmt.Errorf("%w: %s", NetworkErr, err.Error())
	if !IsNetworkError(err) {
		t.Error("err is not network error")
	}
}
