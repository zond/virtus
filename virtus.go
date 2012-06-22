package virtus

/*
 * Core
 */

type Port interface {
	Receive() (Message, bool)
	Send(Message)
}

type Finder interface {
	Find(string) Port
}

type Persistence interface {
	Get(string) ([]byte, bool)
	Set(string, []byte)
	Del(string)
	GetMembers(string) [][]byte
	SetMember(string, []byte)
}

type Thing interface{}

type Hash map[string]Thing

type Message struct {
	Payload Thing
	ReturnPath Port
}

/*
 * Utility
 */

type ChannelPort chan Message
func (self ChannelPort) Receive() (Message, bool) {
	m, ok := <- self
	return m, ok
}
func (self ChannelPort) Send(m Message) {
	self <- m
}

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

const MOVE = "Move"

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
