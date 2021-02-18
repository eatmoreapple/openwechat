package openwechat

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

// Http请求的响应结构体封装
type ReturnResponse struct {
	*http.Response
	err error
}

// Constructor for ReturnResponse
func NewReturnResponse(response *http.Response, err error) *ReturnResponse {
	return &ReturnResponse{Response: response, err: err}
}

// 获取当前请求的错误
func (r *ReturnResponse) Err() error {
	return r.err
}

// json序列化
func (r *ReturnResponse) ScanJSON(v interface{}) error {
	if data, err := r.ReadAll(); err != nil {
		return err
	} else {
		return json.Unmarshal(data, v)
	}
}

// xml序列化
func (r *ReturnResponse) ScanXML(v interface{}) error {
	if data, err := r.ReadAll(); err != nil {
		return err
	} else {
		return xml.Unmarshal(data, v)
	}
}

// 读取请求体
func (r *ReturnResponse) ReadAll() ([]byte, error) {
	if r.Err() != nil {
		return nil, r.Err()
	}
	return ioutil.ReadAll(r.Body)
}
