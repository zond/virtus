package virtus

/*
 * Core
 */

type Thing interface{}

type Message struct {
	Payload Thing
	ReturnPath Port
}

type Port chan Message

type Finder interface {
	Find(string) Port
	Persist(string, Thing)
}

/*
 * Utility
 */

const ROOT = "root"
const CHILDREN_FORMAT = "%s.c"
const USER_ID_FORMAT = "u:%s"

const LOGIN = "Login"
const QUIT = "Quit"
const STRING = "s"
const ACTION = "action"
const ACTIONS = "actions"
const DESC = "desc"

const USERNAME = "Username"
const PASSWORD = "Password"

type Hash map[string]Thing

type Param struct {
	Name string
	Type string
}

func (self Param) Validate(t Thing) bool {
	if self.Type == "s" {
		_, ok := t.(string)
		return ok
	} else if self.Type == "i" {
		_, ok := t.(int)
		return ok
	} else if self.Type == "f" {
		_, ok := t.(float64)
		return ok
	}
	return false
}

type Params []Param

func (self Params) Validate(a Action) bool {
	if len(a.Params) != len(self) {
		return false
	}
	for index, p := range self {
		if !p.Validate(a.Params[index]) {
			return false
		}
	}
	return true
}

type ActionSpec struct {
	Name   string
	Params Params
}

func (self ActionSpec) Validate(a Action) bool {
	return a.Name == self.Name && self.Params.Validate(a)
}

type ActionSpecs []ActionSpec

func (self ActionSpecs) Validate(a Action) bool {
	for _, s := range self {
		if s.Validate(a) {
			return true
		}
	}
	return false
}

type Query struct {
	Desc        string
	ActionSpecs ActionSpecs
}

func (self Query) Validate(a Action) bool {
	return self.ActionSpecs.Validate(a)
}

type Action struct {
	Name   string
	Params []Thing
}

type ChannelPort chan Message
func (self ChannelPort) Receive(m *Message) bool {
	if x, ok := <- self; ok {
		(*m).Payload = x.Payload
		(*m).ReturnPath = x.ReturnPath
		return ok
	}
	return nil, false
}
func (self ChannelPort) Send(m Message) {
	self <- m
}

type Portal struct {
	Port
}
func (self *Portal) Query(q Query) (a Action, ok bool) {
	for !q.Validate(a) {
		a = Action{}
		self.Port.Send(q)
		if ok = self.Port.Receive(&a); !ok {
			return nil, false
		}
	}
	return a, true
}

