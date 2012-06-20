
package main

import (
        "net/http"
        "code.google.com/p/go.net/websocket"
	"log"
	"fmt"
	"io"
)

type EOF string

func (self *object) webServe() {
        err := http.ListenAndServe(":8080", websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(EOF); !ok {
					log.Println(ws, ": ", r)
				}
			}
			ws.Close()
		}()
		(&conn{ws, self}).authenticate()
	}))
        if err != nil {
                panic("While starting web socket server: " + err.Error())
        }
}

type conn struct {
	conn *websocket.Conn
	root *object
}
	
func (self *conn) send(m mess) {
	if err := websocket.JSON.Send(self.conn, m); err != nil {
		message := fmt.Sprint("While trying to send ", m, " to ", self.conn, ": ", err)
		if err == io.EOF {
			panic(EOF(message))
		} else {
			panic(message)
		}
	}
}
func (self *conn) recv(r *resp) {
	if err := websocket.JSON.Receive(self.conn, r); err != nil {
		message := fmt.Sprint("While trying to receive from ", self.conn, ": ", err)
		if err == io.EOF {
			panic(EOF(message))
		} else {
			panic(message)
		}
	}
}
func (self *conn) query(m mess) resp {
	var r resp
	for !m.validate(r) {
		self.send(m)
		self.recv(&r)
	}
	return r
}
func (self *conn) authenticate() {
	var home *object
	for home == nil {
		resp := self.query(
			mess{"Enter username and password", 
			actions{action{LOGIN, params{ param{USERNAME, STRING}, param{PASSWORD, STRING}}},
				action{QUIT, params{}}}})
		if resp.Action == LOGIN {
			if o := self.root.createChild(fmt.Sprintf(USER_ID_FORMAT, resp.Params[0].(string))).load(); o.fresh {
				o.setPassword(resp.Params[1].(string))
				self.send(mess{Desc: fmt.Sprint("Created new account ", resp.Params[0])})
				break
			} else {
				if o.authenticate(resp.Params[1].(string)) {
					self.send(mess{Desc: fmt.Sprint("Authenticated as ", resp.Params[0])})
					break
				} else {
					self.send(
						mess{"Bad username or password", 
						actions{action{LOGIN, params{param{USERNAME, STRING}, param{PASSWORD, STRING}}},
							action{QUIT, params{}}}})
				}
			}
		} else if resp.Action == QUIT {
			break
		}
	}
}

