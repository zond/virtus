
package main

import (
        "net/http"
        "code.google.com/p/go.net/websocket"
	"encoding/json"
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
		(&conn{ws, json.NewDecoder(ws), self}).start()
	}))
        if err != nil {
                panic("While starting web socket server: " + err.Error())
        }
}

type conn struct {
	conn *websocket.Conn
	decoder *json.Decoder
	root *object
}
	
func (self *conn) send(t thing) {
	if err := websocket.JSON.Send(self.conn, t); err != nil {
		message := fmt.Sprint("While trying to send ", t, " to ", self.conn, ": ", err)
		if err == io.EOF {
			panic(EOF(message))
		} else {
			panic(message)
		}
	}
}
func (self *conn) recv(t thing) {
	if err := self.decoder.Decode(t); err != nil {
		message := fmt.Sprint("While trying to receive from ", self.conn, ": ", err)
		if err == io.EOF {
			panic(EOF(message))
		} else {
			panic(message)
		}
	}
}
func (self *conn) query(m query) mess {
	var r mess
	for !m.validate(r) {
		self.send(m)
		self.recv(&r)
	}
	return r
}
func (self *conn) start() {
	obj := self.authenticate()
	if obj == nil {
		return
	}
	self.connect(obj)
}
func (self *conn) connect(o *object) {
	for {
		var got thing
		self.recv(&got)
		o.port <- got
	}
}
func (self *conn) authenticate() *object {
	for {
		mess := self.query(
			query{"Enter username and password", 
			actions{action{LOGIN, params{ param{USERNAME, STRING}, param{PASSWORD, STRING}}},
				action{QUIT, params{}}}})
		if mess.Action == LOGIN {
			if o := self.root.createChild(fmt.Sprintf(USER_ID_FORMAT, mess.Params[0].(string))).load(); o.fresh {
				o.setPassword(mess.Params[1].(string))
				self.send(query{Desc: fmt.Sprint("Created new account ", mess.Params[0])})
				o.boot()
				return o
			} else {
				if o.authenticate(mess.Params[1].(string)) {
					self.send(query{Desc: fmt.Sprint("Authenticated as ", mess.Params[0])})
					o.boot()
					return o
				} else {
					self.send(
						query{"Bad username or password", 
						actions{action{LOGIN, params{param{USERNAME, STRING}, param{PASSWORD, STRING}}},
							action{QUIT, params{}}}})
				}
			}
		} else if mess.Action == QUIT {
			break
		}
	}
	return nil
}

