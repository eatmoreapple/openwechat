package openwechat

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func ToBuffer(v interface{}) (*bytes.Buffer, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(buf), nil
}

func GetRandomDeviceId() string {
	rand.Seed(time.Now().Unix())
	str := ""
	for i := 0; i < 15; i++ {
		r := rand.Intn(9)
		str += strconv.Itoa(r)
	}
	return "e" + str
}

func getWebWxDataTicket(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "webwx_data_ticket" {
			return cookie.Value
		}
	}
	return ""
}
