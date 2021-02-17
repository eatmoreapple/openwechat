package openwechat

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

//func getSendMessageError(body io.Reader) error {
//	data, err := ioutil.ReadAll(body)
//	if err != nil {
//		return err
//	}
//	var item struct{ BaseResponse BaseResponse }
//	if err = json.Unmarshal(data, &item); err != nil {
//		return err
//	}
//	if !item.BaseResponse.Ok() {
//		return errors.New(item.BaseResponse.ErrMsg)
//	}
//	return nil
//}

func getWebWxDataTicket(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "webwx_data_ticket" {
			return cookie.Value
		}
	}
	return ""
}

func getUpdateMember(resp *http.Response, err error) (Members, error) {
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var item struct {
		BaseResponse BaseResponse
		ContactList  Members
	}
	if err = json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	if !item.BaseResponse.Ok() {
		return nil, item.BaseResponse
	}
	return item.ContactList, nil
}

func getResponseBody(resp *http.Response) ([]byte, error) {
	if data, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

