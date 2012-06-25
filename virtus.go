
package main

import (
	"github.com/zond/cabinet"
	"fmt"
	"encoding/gob"
	"bytes"
)

const (
	BOOT = "boot"
)

type port chan message

type ports map[string]port

type thing interface{}

type message struct {
	typ string
	payload thing
	caller port
}

type object struct {
	port port
	neighbors ports
}
func newObject() *object {
	return &object{}
}
func (self *object) boot() {
	self.port = make(port)
	self.neighbors = make(ports)
	go self.listen()
	self.port <- message{BOOT, nil, self.port}
}
func (self *object) listen() {
	for message := range self.port {
		fmt.Println(self.port, " got ", message)
	}
}

type loader struct {
	objects map[string]*object
	cabinet *cabinet.KCDB
}
func newLoader() *loader {
	return &loader{make(map[string]*object), cabinet.New()}
}
func (self *loader) load() {
	self.cabinet.Open("objects.kch", cabinet.KCOWRITER | cabinet.KCOCREATE)
	self.bootObject("root")
}
func (self *loader) bootObject(name string) {
	object := self.getObject(name)
	object.boot()
}
func (self *loader) loadObject(name string) *object {
	if b, err := self.cabinet.Get([]byte(name)); err == nil {
		decoder := gob.NewDecoder(bytes.NewBuffer(b))
		rval := &object{}
		if err = decoder.Decode(rval); err != nil {
			panic(fmt.Sprint("Unable to load ", string(b), " into ", rval, ": ", err))
		}
		return rval
	}
	return nil
}
func (self *loader) getObject(name string) *object {
	fmt.Println("loading ", name)
	if object, ok := self.objects[name]; ok {
		return object
	}
	if rval := self.loadObject(name); rval != nil {
		return rval
	}
	return self.createObject(name)
}
func (self *loader) createObject(name string) *object {
	rval := &object{}
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(rval); err != nil {
		panic(fmt.Sprint("Unable to encode ", rval, " using ", encoder, ": ", err))
	}
	if err := self.cabinet.Set([]byte(name), buffer.Bytes()); err != nil {
		panic(fmt.Sprint("Unable to store ", rval, " in ", self.cabinet, ": ", err))
	}
	return rval
}

func main() {
	l := newLoader()
	l.load()
}