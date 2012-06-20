
package main

import (
        "net/http"
        "code.google.com/p/go.net/websocket"
	"log"
)

const LOGIN = "Login"
const QUIT = "Quit"
const STRING = "s"
const ACTION = "action"
const ACTIONS = "actions"
const DESC = "desc"

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
	defer ws.Close()
	m := hash{
		DESC: "Enter username and password", 
		ACTIONS : hash{
			LOGIN : ary{ ary{ "Username", STRING }, ary{ "Password", STRING }},
			QUIT : ary {},
		},
	}
	if err := websocket.JSON.Send(ws, m); err == nil {
		var home *object
		var resp hash
		for home == nil {
			if err := websocket.JSON.Receive(ws, &resp); err == nil {
				if resp[ACTION] == LOGIN {
					log.Println("login from ", resp)
				} else if resp[ACTION] == QUIT {
					break
				}
			} else {
				break
			}
		}
	} else {
		log.Println("While trying to send ", m, " to ", ws, ": ", err)
	}
}

