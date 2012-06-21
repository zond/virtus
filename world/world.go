package main

import (
	. "../"
	"../persistence"
	"../webserver"
)

type persistenceFinder struct {
	Persistence
}

func (self *persistenceFinder) Find(s string) Object {
	return findObject(self, s)
}
func (self *persistenceFinder) Create(s, p string) Object {
	o := createObject(self, s)
	o.setPassword(p)
	return o
}

func main() {
	webserver.Serve(&persistenceFinder{persistence.New()})
}
