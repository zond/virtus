
package main

import (
	"github.com/zond/tools"
	"github.com/simonz05/godis"
	"fmt"
	"encoding/json"
)

const root = "root"
const children_format = "%s.c"

type thing interface{}
type hash map[string]thing
type ary []thing

type port chan thing

var redis = godis.New("", 0, "")

type object struct {
	id string
	port port
	parent port
	children map[string]port
	state hash
}
func createObject(id string) *object {
	return &object{id: id, port: make(port), children: make(map[string]port)}
}
func getRoot() *object {
	rval := createObject(root)
	rval.load()
	return rval
}
func (self *object) createChild() *object {
	rval := self.loadChild(tools.Uuid())
	return rval
}
func (self *object) load() {
	elem, err := redis.Get(self.id)
	if err == nil {
		json.Unmarshal(elem, &(self.state))
		self.loadChildren()
	} else if err.Error() == "Nonexisting key" {
		self.state = make(hash)
		self.save()
	} else {
		panic(fmt.Sprint("Unable to load ", self.id, ": ", err))
	}
	go self.run()
}
func (self *object) loadChildren() {
	keys, err := redis.Smembers(fmt.Sprintf(children_format, self.id))
	if err != nil {
		panic(fmt.Sprint("Unable to load children of ", self, ": ", err))
	}
	for _, reply := range keys.Elems {
		self.loadChild(string(reply.Elem))
	}
}
func (self *object) loadChild(id string) *object {
	rval := createObject(id)
	rval.parent = self.port
	self.children[rval.id] = rval.port
	rval.load()
	return rval
}
func (self *object) run() {
	for t := range self.port {
		fmt.Println("got", t)
	}
}
func (self *object) save() {
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
	root := getRoot()
	newWebServer(root)
}
