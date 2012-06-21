package webserver

import (
	. "../"
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net/http"
)

type EOF string

type message struct {
	payload    Thing
	returnPath chan Message
}

func (self *message) Payload() Thing {
	return self.payload
}
func (self *message) ReturnPath() chan Message {
	return self.returnPath
}

func Serve(f Finder) {
	err := http.ListenAndServe(":8080", websocket.Handler(func(ws *websocket.Conn) {
		(&conn{ws}).start(f)
	}))
	if err != nil {
		panic("While starting web socket server: " + err.Error())
	}
}

type conn struct {
	*websocket.Conn
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
func (self *conn) query(q Query) Action {
	var a Action
	for !q.Validate(a) {
		self.send(q)
		self.recv(&a)
	}
	return a
}
func (self *conn) start(f Finder) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(EOF); !ok {
				panic(r)
			}
		}
		self.Close()
	}()
	r := self.authenticate(f)
	if r == nil {
		return
	}
	self.serve(r)
}
func (self *conn) listen(c chan Message) {
	for m := range c {
		self.send(m.Payload())
	}
}
func (self *conn) serve(o Object) {
	returnPath := make(chan Message)
	go self.listen(returnPath)
	o.Start()
	for {
		var got Thing
		self.recv(&got)
		o.Port().Send(&message{got, returnPath})
	}
}
func (self *conn) authenticate(f Finder) Object {
	for {
		mess := self.query(
			Query{"Enter username and password",
				ActionSpecs{ActionSpec{LOGIN, Params{Param{USERNAME, STRING}, Param{PASSWORD, STRING}}},
					ActionSpec{QUIT, Params{}}}})
		if mess.Name == LOGIN {
			if o := f.Find(mess.Params[0].(string)); o == nil {
				o = f.Create(mess.Params[0].(string), mess.Params[1].(string))
				self.send(Query{Desc: fmt.Sprint("Created new account ", mess.Params[0])})
				return o
			} else {
				if o.Authenticate(mess.Params[1].(string)) {
					self.send(Query{Desc: fmt.Sprint("Authenticated as ", mess.Params[0])})
					return o
				} else {
					self.send(
						Query{"Bad username or password",
							ActionSpecs{ActionSpec{LOGIN, Params{Param{USERNAME, STRING}, Param{PASSWORD, STRING}}},
								ActionSpec{QUIT, Params{}}}})
				}
			}
		} else if mess.Name == QUIT {
			break
		}
	}
	return nil
}
