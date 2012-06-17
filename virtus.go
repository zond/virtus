
package main

import (
	"github.com/zond/tools"
	"github.com/simonz05/godis"
	"fmt"
	"encoding/json"
)

const root = "root"
const children_format = "%s.c"

type Thing interface{}

type Port chan Thing

var redis = godis.New("", 0, "")

type Object struct {
	id string
	port Port
	parent Port
	children map[string]Port
	state map[Thing]Thing
}
func CreateObject(id string) *Object {
	return &Object{id: id, port: make(Port), children: make(map[string]Port)}
}
func GetRoot() *Object {
	rval := CreateObject(root)
	rval.Load()
	return rval
}
func (self *Object) CreateChild() *Object {
	rval := self.LoadChild(tools.Uuid())
	return rval
}
func (self *Object) Load() {
	elem, err := redis.Get(self.id)
	if err == nil {
		json.Unmarshal(elem, &(self.state))
		self.LoadChildren()
	} else if err.Error() == "Nonexisting key" {
		self.state = make(map[Thing]Thing)
		self.Save()
	} else {
		panic(fmt.Sprint("Unable to load ", self.id, ": ", err))
	}
	go self.Run()
}
func (self *Object) LoadChildren() {
	keys, err := redis.Smembers(fmt.Sprintf(children_format, self.id))
	if err != nil {
		panic(fmt.Sprint("Unable to load children of ", self, ": ", err))
	}
	for _, reply := range keys.Elems {
		self.LoadChild(string(reply.Elem))
	}
}
func (self *Object) LoadChild(id string) *Object {
	rval := CreateObject(id)
	rval.parent = self.port
	self.children[rval.id] = rval.port
	rval.Load()
	return rval
}
func (self *Object) Run() {
	for t := range self.port {
		fmt.Println("got", t)
	}
}
func (self *Object) Save() {
	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Sprint("Unable to marshal ", self, ": ", err))
	}
	err = redis.Set(self.id, b)
	if err != nil {
		panic(fmt.Sprint("Unable to store ", self, ": ", err))
	}
	_, err = redis.Del(fmt.Sprintf(children_format, self.id))
	if err != nil {
		panic(fmt.Sprint("Unable to clear children from ", self, ": ", err))
	}
	for id, _ := range self.children {
		_, err := redis.Sadd(fmt.Sprintf(children_format, self.id), id)
		if err != nil {
			panic(fmt.Sprintf("Unable to add ", id, " to children of ", self, ": ", err))
		}
	}
}

func main() {
	root := GetRoot()
	fmt.Println(root)
}
