
package main

import (
	"encoding/json"
        "code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"os"
	"log"
)

func receive(ws *websocket.Conn, c chan string) {
	var d interface{}
	for {
		if err := websocket.JSON.Receive(ws, &d); err == nil {
			if s, err := json.MarshalIndent(d, "", " "); err == nil {
				c <- string(s)
			} else {
				log.Println("While trying to marshal ", d, ": ", err)
			}
		} else {
			if err.Error() == "EOF" {
				close(c)
				break
			} else {
				log.Println("While trying to receive: ", err)
			}
		}
	}
}

func read(c chan interface{}) {
	var d interface{}
	decoder := json.NewDecoder(os.Stdin)
	for {
		if err := decoder.Decode(&d); err == nil {
			c <- d
		} else {
			log.Println("While trying to parse from os.Stdin: ", err)
			decoder = json.NewDecoder(os.Stdin)
		}
	}		
}

func main() {
	flag.Parse()
	if flag.NArg() == 1 {
		origin := "http://localhost/"
		url := flag.Arg(0)
		ws, err := websocket.Dial(url, "", origin)
		if err != nil {
			panic(fmt.Sprint("While trying to connect to ", url, ": ", err))
		}
		in := make(chan string)
		out := make(chan interface{})
		go receive(ws, in)
		go read(out)
		for {
 			select {
			case s, ok := <- in:
				if ok {
					fmt.Println(s)
				} else {
					return
				}
			case d := <- out:
				websocket.JSON.Send(ws, d)
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "Usage: cli URL")
	}
}
