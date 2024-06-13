package cache

import (
	"bytes"
	"encoding/gob"
	"github.com/goccy/go-json"
)

// Codec is an interface that allows encoding and decoding of byte slices.
type Codec interface {
	Encode(value interface{}) (data []byte, err error)
	Decode(data []byte, value interface{}) (err error)
}

var _ Codec = (*JsonCodec)(nil)

type JsonCodec struct{}

func (c *JsonCodec) Encode(value interface{}) (data []byte, err error) {
	return json.Marshal(value)
}

func (c *JsonCodec) Decode(data []byte, value interface{}) (err error) {
	return json.Unmarshal(data, value)
}

var _ Codec = (*GobCodec)(nil)

type GobCodec struct{}

func (c *GobCodec) Encode(value interface{}) (data []byte, err error) {
	var w bytes.Buffer
	err = gob.NewEncoder(&w).Encode(value)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (c *GobCodec) Decode(data []byte, value interface{}) (err error) {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(value)
}
