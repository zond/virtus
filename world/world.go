package main

import (
	. "../"
	"../persistence"
	"../webserver"
)

type simpleFinder struct {
	Persistence
}

func (self *simpleFinder) Find(s string) Port {
	return findObject(self, s)
}

func main() {
	webserver.Serve(&simpleFinder{persistence.New()})
}
