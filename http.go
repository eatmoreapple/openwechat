package openwechat

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type ReturnResponse struct {
	*http.Response
	err error
}

func NewReturnResponse(response *http.Response, err error) *ReturnResponse {
	return &ReturnResponse{Response: response, err: err}
}

func (r *ReturnResponse) Err() error {
	return r.err
}

func (r *ReturnResponse) ScanJSON(v interface{}) error {
	if data, err := r.ReadAll(); err != nil {
		return err
	} else {
		return json.Unmarshal(data, v)
	}
}

func (r *ReturnResponse) ScanXML(v interface{}) error {
	if data, err := r.ReadAll(); err != nil {
		return err
	} else {
		return xml.Unmarshal(data, v)
	}
}

func (r *ReturnResponse) ReadAll() ([]byte, error) {
	if r.Err() != nil {
		return nil, r.Err()
	}
	return ioutil.ReadAll(r.Body)
}
