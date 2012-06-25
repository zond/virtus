package main

import (
	"../webserver"
)


func main() {
	webserver.Serve(newDirectory())
}
