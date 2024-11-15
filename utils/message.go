package utils

import "encoding/json"

var (
	Magic    = []byte{0x09, 0x03, 0x07, 0x01}
	MagicLen = 4
	V1       = 1
)

const (
	StatusSuccess = 200
)

type Message interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Hello struct {
	Version int    `json:"version"`
	Key     string `json:"key"`
	Nonce   string `json:nonce"`
}

func (self *Hello) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *Hello) Unmarshal(data []byte) error {
	return json.Unmarshal(data, self)
}

type HelloResp struct {
	Status int `json:"status"`
}

func (self *HelloResp) Marshal() ([]byte, error) {
	return json.Marshal(self)
}

func (self *HelloResp) Unmarshal(data []byte) error {
	return json.Unmarshal(data, self)
}

func Marshal(msg Message) ([]byte, error) {
	data, err := msg.Marshal()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Unmarshal(data []byte, msg Message) error {
	return msg.Unmarshal(data)
}
