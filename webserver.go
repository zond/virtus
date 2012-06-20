
package main

import (
        "net/http"
        "code.google.com/p/go.net/websocket"
	"fmt"
)

type webServer struct {
	root *object
}
func newWebServer(r *object) *webServer {
	rval := &webServer{r}
        err := http.ListenAndServe(":8080", websocket.Handler(func(ws *websocket.Conn) {
		rval.handle(ws)
	}))
        if err != nil {
                panic("While starting web socket server: " + err.Error())
        }
	return rval
}
func (self *webServer) handle(ws *websocket.Conn) {
	m := hash{
		"desc": "Enter username and password", 
		"actions" : hash{
			"Login" : ary{ ary{ "Username", "s" }, ary{ "Password", "s" }},
			"Quit" : ary {},
		},
	}
	if err := websocket.JSON.Send(ws, m); err == nil {
		var home *object
		var resp interface{}
		for home == nil {
			if err := websocket.JSON.Receive(ws, &resp); err == nil {
				fmt.Println("got ", resp)
			}
		}
	}
}

