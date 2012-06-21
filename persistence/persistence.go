package persistence

import (
	. "../"
	"github.com/simonz05/godis"
)

func New() Persistence {
	return &persistence{godis.New("", 0, "")}
}

type persistence struct {
	*godis.Client
}

func (self persistence) Get(s string) ([]byte, bool) {
	rval, err := self.Client.Get(s)
	if err != nil {
		if err.Error() == "Nonexisting key" {
			return nil, false
		} else {
			panic(err.Error())
		}
	}
	return rval, true
}
func (self persistence) GetMembers(s string) [][]byte {
	keys, err := self.Client.Smembers(s)
	if err != nil {
		panic(err.Error())
	}
	var rval [][]byte
	for _, reply := range keys.Elems {
		rval = append(rval, reply.Elem)
	}
	return rval
}
func (self persistence) Set(k string, v []byte) {
	if err := self.Client.Set(k, v); err != nil {
		panic(err.Error())
	}
}
func (self persistence) Del(s string) {
	if _, err := self.Client.Del(s); err != nil {
		panic(err.Error())
	}
}
func (self persistence) SetMember(s string, b []byte) {
	if _, err := self.Client.Sadd(s, b); err != nil {
		panic(err.Error())
	}
}
