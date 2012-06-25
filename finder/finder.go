
package finder

import (
	. "../"
	"../object"
	"github.com/zond/cabinet"
)

const (
	invalid = iota
	FIND
	PERSIST
)

func NewFinder() *finder {
	kc := cabinet.New()
	rval := &finder{kc.Open("virtus.kch", cabinet.KCOWRITER | cabinet.KCOCREATE), make(map[string]Port), make(chan request)}
	go rval.serve()
	return rval
}

type request struct {
	key string
	value Thing
	typ int
	returnPath chan Port
}
type finder struct {
	cabinet *cabinet.KCDB
	ports map[string]Port
	requests chan request
}
func (self *finder) serve() {
	for request := range self.requests {
		switch request.typ {
		case FIND:
			if port, ok := self.ports[request.key]; ok {
				request.returnPath <- port
			} else {
				obj := object.Create(request.key)
				obj.Boot(self)
				self.ports[request.key] = obj.Port
				request.returnPath <- obj.Port
			}
		case PERSIST:
			buffer := &bytes.Buffer{}
			encoder := gob.NewEncoder(buffer)
			encoder.Encode(request.value)
			self.cabinet.Set([]byte(request.key), buffer.Bytes())
			close(request.returnPath)
		}
	}
}
func (self *directory) Find(s string) Port {
	request := Request{key: s, returnPath: make(chan Port), typ: FIND}
	self.requests <- request
	return <-request.returnPath
}
func (self *