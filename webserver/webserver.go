package webserver

import (
	. "../"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net/http"
)

type EOF string

func Serve(r Finder) {
	err := http.ListenAndServe(":8080", websocket.Handler(func(ws *websocket.Conn) {
		(&conn{ws, Portal(make(ChannelPort))}).start(r)
	}))
	if err != nil {
		panic("While starting web socket server: " + err.Error())
	}
}

type conn struct {
	conn *websocket.Conn
	port *Portal
}

func (self *conn) send(t Thing) {
	if err := websocket.JSON.Send(self.Conn, t); err != nil {
		message := fmt.Sprint("While trying to send ", t, " to ", self.Conn, ": ", err)
		if err == io.EOF {
			panic(EOF(message))
		} else {
			panic(message)
		}
	}
}
func (self *conn) recv(t Thing) {
	if err := websocket.JSON.Receive(self.Conn, t); err != nil {
		message := fmt.Sprint("While trying to receive from ", self.Conn, ": ", err)
		if err == io.EOF {
			panic(EOF(message))
		} else {
			panic(message)
		}
	}
}
func (self *conn) start(r Finder) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(EOF); !ok {
				panic(r)
			}
		}
		self.Close()
	}()
	self.serve(self.find(r))
}
func (self *conn) listen() {
	for {
		m := Message{}
		if ok := self.port.Receive(&m); !ok {
			break
		}
		self.send(m.Payload)
	}
}
func (self *conn) serve(p Port) {
	go self.listen()
	for {
		var got Thing
		self.recv(&got)
		p.Send(Message{got, self.port})
	}
}
func (self *conn) find(r Finder) Port {
	for {
		mess := self.query(
			Query{"Login, Create account or Quit",
				ActionSpecs{ActionSpec{LOGIN, Params{Param{USERNAME, STRING}, Param{PASSWORD, STRING}}},
				            ActionSpec{CREATE, Params{Param{USERNAME, STRING}, Param{PASSWORD, STRING}}},
					    ActionSpec{QUIT, Params{}}}})
		if mess.Name != QUIT {
			p := r.Find(fmt.Sprintf(USER_ID_FORMAT, mess.Params[0].(string)))
			p.Send(Message{mess, self.port})
			return p
		} else { 
			break
		}
	}
	return nil
}
