package openwechat

import (
	"encoding/json"
	"io"
)

// Serializer is an interface for encoding and decoding data.
type Serializer interface {
	Encode(writer io.Writer, v interface{}) error
	Decode(reader io.Reader, v interface{}) error
}

// JsonSerializer is a serializer for json.
type JsonSerializer struct{}

// Encode encodes v to writer.
func (j JsonSerializer) Encode(writer io.Writer, v interface{}) error {
	return json.NewEncoder(writer).Encode(v)
}

// Decode decodes data from reader to v.
func (j JsonSerializer) Decode(reader io.Reader, v interface{}) error {
	return json.NewDecoder(reader).Decode(v)
}
