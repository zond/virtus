
package main

import (
	"github.com/simonz05/godis"
	"fmt"
	"encoding/json"
)

type object struct {
	id string
	port port
	parent port
	children map[string]port
	state hash
	fresh bool
	redis *godis.Client
}
func createObject(redis *godis.Client, id string, parent *object) *object {
	rval := &object{id: id, port: make(port), children: make(map[string]port), redis: redis}
	if parent != nil {
		rval.parent = parent.port
		parent.children[rval.id] = rval.port
	}
	return rval
}
func loadAndBootRoot(redis *godis.Client) *object {
	return createObject(redis, ROOT, nil).loadAndBoot()
}
func (self *object) createChild(id string) *object {
	return createObject(self.redis, id, self)
}
func (self *object) loadAndBootChild(id string) *object {
	return createObject(self.redis, id, self).loadAndBoot()
}
func (self *object) loadAndBoot() *object {
	self.load()
	go self.run()
	self.loadAndBootChildren()
	return self
}
func (self *object) load() *object {
	elem, err := self.redis.Get(self.id)
	if err == nil {
		json.Unmarshal(elem, &(self.state))
		self.fresh = false
	} else if err.Error() == "Nonexisting key" {
		self.state = make(hash)
		self.fresh = true
	} else {
		panic(fmt.Sprint("Unable to load ", self.id, ": ", err))
	}
	return self
}
func (self *object) loadAndBootChildren() *object {
	keys, err := self.redis.Smembers(fmt.Sprintf(CHILDREN_FORMAT, self.id))
	if err != nil {
		panic(fmt.Sprint("Unable to load children of ", self, ": ", err))
	}
	for _, reply := range keys.Elems {
		self.loadAndBootChild(string(reply.Elem))
	}
	return self
}
func (self *object) run() {
	for t := range self.port {
		fmt.Println("got", t)
	}
}
func (self *object) save() *object {
	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Sprint("Unable to marshal ", self, ": ", err))
	}
	err = self.redis.Set(self.id, b)
	if err != nil {
		panic(fmt.Sprint("Unable to store ", self, ": ", err))
	}
	_, err = self.redis.Del(fmt.Sprintf(CHILDREN_FORMAT, self.id))
	if err != nil {
		panic(fmt.Sprint("Unable to clear children from ", self, ": ", err))
	}
	for id, _ := range self.children {
		_, err := self.redis.Sadd(fmt.Sprintf(CHILDREN_FORMAT, self.id), id)
		if err != nil {
			panic(fmt.Sprintf("Unable to add ", id, " to children of ", self, ": ", err))
		}
	}
	return self
}

