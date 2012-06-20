package main

import (
	"github.com/simonz05/godis"
)

type thing interface{}
type hash map[string]thing
type ary []thing

type port chan thing

type param struct {
	Name string
	Type string
}
func (self param) validate(t thing) bool {
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
type params []param
func (self params) validate(r mess) bool {
	if len(r.Params) != len(self) {
		return false
	}
	for index, p := range self {
		if !p.validate(r.Params[index]) {
			return false
		}
	}
	return true
}

type action struct {
	Name string
	Params params
}
func (self action) validate(r mess) bool {
	return r.Action == self.Name && self.Params.validate(r)
}
type actions []action
func (self actions) validate(r mess) bool {
	for _, a := range(self) {
		if a.validate(r) {
			return true
		}
	}
	return false
}

type query struct {
	Desc string
	Actions actions
}
func (self query) validate(r mess) bool {
	return self.Actions.validate(r)
}

type mess struct {
	Action string
	Params []thing
}

func main() {
	loadAndBootRoot(godis.New("", 0, "")).webServe()
}
