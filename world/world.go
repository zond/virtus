package main

import (
	"../persistence"
	"../webserver"
	. "../"
)

type persistenceFinder struct {
	Persistence
}
func (self *persistenceFinder) Find(s string) (Object, error) {
	o, err := findObject(self, s)
	if err != nil {
		return nil, err
	}
	if o != nil {
		return o, nil
	}
	return nil, nil
}
func (self *persistenceFinder) Create(s, p string) Object {
	o := createObject(self, s)
	o.setPassword(p)
	return o
}

func main() {
	webserver.Serve(&persistenceFinder{persistence.New()})
}

