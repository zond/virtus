package main

import (
	"crypto/sha1"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	. "../"
	"github.com/zond/tools"
)

type port chan Message
func (self port) Receive() Message {
	return <- self
}
func (self port) Send(m Message) {
	self <- m
}

type object struct {
	id       string
	password string
	salt     string
	port     Port
	parent   Port
	children map[string]Port
	state    Hash
	fresh    bool
	persistence Persistence
}
func loadObject(p Persistence, id string, data []byte) (*object, error) {
	rval := createObject(p, id)
	if err := json.Unmarshal(data, rval); err != nil {
		return nil, err
	}
	return rval, nil
}
func findObject(p Persistence, id string) (*object, error) {
	data, err := p.Get(id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return loadObject(p, id, data)
}
func createObject(p Persistence, id string) *object {
	return &object{id: id, port: make(port), children: make(map[string]Port), persistence: p}
}
func (self *object) Start() error {
	if err := self.startChildren(); err != nil {
		return err
	}
	return nil
}
func (self *object) Stop() {
	close(self.port.(port))
}
func (self *object) Authenticate(password string) bool {
	h := sha1.New()
	h.Write([]byte(self.salt))
	return subtle.ConstantTimeCompare(h.Sum([]byte(password)), tools.NewBigIntString(self.password, tools.MAX_BASE).Bytes()) == 1
}
func (self *object) Port() Port {
	return self.port
}


func (self *object) setPassword(password string) {
	self.salt = tools.Uuid()
	h := sha1.New()
	h.Write([]byte(self.salt))
	self.password = tools.NewBigIntBytes(h.Sum([]byte(password))).BaseString(tools.MAX_BASE)
}
func (self *object) startChildren() error {
	ary, err := self.persistence.GetMembers(fmt.Sprintf(CHILDREN_FORMAT, self.id))
	if err != nil {
		return err
	}
	for _, data := range ary {
		child, err := findObject(self.persistence, string(data))
		if err != nil {
			return err
		}
		if child != nil {
			child.parent = self.port
			self.children[child.id] = child.port
			child.Start()
		}
	}
	return nil
}
func (self *object) run() {
	for m := range self.port.(port) {
		fmt.Printf("%v got %v of type %T\n", self.id, m, m)
	}
}
func (self *object) save() *object {
	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Sprint("Unable to marshal ", self, ": ", err))
	}
	err = self.persistence.Set(self.id, b)
	if err != nil {
		panic(fmt.Sprint("Unable to store ", self, ": ", err))
	}
	_, err = self.persistence.Del(fmt.Sprintf(CHILDREN_FORMAT, self.id))
	if err != nil {
		panic(fmt.Sprint("Unable to clear children from ", self, ": ", err))
	}
	for id, _ := range self.children {
		if err := self.persistence.SetMember(fmt.Sprintf(CHILDREN_FORMAT, self.id), []byte(id)); err != nil {
			panic(fmt.Sprintf("Unable to add ", id, " to children of ", self, ": ", err))
		}
	}
	return self
}
