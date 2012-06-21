
package persistence

import (
	"github.com/simonz05/godis"
	. "../"
)

func New() Persistence {
	return &persistence{godis.New("", 0, "")}
}

type persistence struct {
	*godis.Client
}
func (self persistence) Get(s string) ([]byte, error) {
	return self.Client.Get(s)
}
func (self persistence) GetMembers(s string) ([][]byte, error) {
	keys, err := self.Client.Smembers(s)
	if err != nil {
		return nil, err
	}
	var rval [][]byte
	for _, reply := range keys.Elems {
		rval = append(rval, reply.Elem)
	}
	return rval, nil
}
func (self persistence) Set(k string, v []byte) error {
	return self.Client.Set(k, v)
}
func (self persistence) Del(s string) (int64, error) {
	return self.Client.Del(s)
}
func (self persistence) SetMember(s string, b []byte) error {
	if _, err := self.Client.Sadd(s, b); err != nil {
		return err
	}
	return nil
}
