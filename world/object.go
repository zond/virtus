package main

import (
	. "../"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"github.com/zond/tools"
)

type port chan Message
func (self port) Receive() Message {
	return <-self
}
func (self port) Send(m Message) {
	self <- m
}

type object struct {
	id          string
	password    string
	salt        string
	port        Port
	parent      Port
	children    map[string]Port
	neighbors   map[string]Port
	state       Hash
	fresh       bool
	persistence Persistence
}

func loadObject(p Persistence, id string, data []byte) *object {
	rval := createObject(p, id)
	if err := json.Unmarshal(data, rval); err != nil {
		panic(err.Error())
	}
	return rval
}
func findObject(p Persistence, id string) *object {
	if data, ok := p.Get(id); ok {
		return loadObject(p, id, data)
	}
	return nil
}
func createObject(p Persistence, id string) *object {
	return &object{id: id, port: make(port), children: make(map[string]Port), neighbors: make(map[string]Port), persistence: p}
}
func (self *object) Start() error {
	if err := self.startChildren(); err != nil {
		return err
	}
	go self.run()
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
	for data := range self.persistence.GetMembers(fmt.Sprintf(CHILDREN_FORMAT, self.id)) {
		if child := findObject(self.persistence, string(data)); child != nil {
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
	self.persistence.Set(self.id, b)
	self.persistence.Del(fmt.Sprintf(CHILDREN_FORMAT, self.id))
	for id, _ := range self.children {
		self.persistence.SetMember(fmt.Sprintf(CHILDREN_FORMAT, self.id), []byte(id))
	}
	return self
}
