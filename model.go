package wserve

import (
	"bytes"
	"encoding/json"
)

//var _ IUser = (*DUser)(nil)

type IUser interface {
	Compare(interface{}) bool
	Major() interface{}
}

type DUser struct {
	Addr string
}

func (u *DUser) Compare(addr interface{}) bool {
	a, err := json.Marshal(u.Major())
	if err != nil {
		return false
	}
	b, err := json.Marshal(addr)
	if err != nil {
		return false
	}
	return bytes.Equal(a, b)

}

func (u *DUser) Major() interface{} {
	return u.Addr
}

type Body struct {
	IFrom   IUser       `json:"-"`
	From    interface{} `json:"from"`
	To      interface{} `json:"to"`
	Message interface{} `json:"message"`
	Type    string      `json:"type"`
	Operate string      `json:"operate"`
}

func (rb *Body) Bytes() []byte {
	marshal, _ := json.Marshal(rb)
	return marshal
}
