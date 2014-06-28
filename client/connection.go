package client

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"os"
	//	"strings"
)

type Connection struct {
	ws       *websocket.Conn
	Status   string
	url      string
	Incoming chan string
}

func (conn *Connection) Start(url string) {
	conn.url = url
	conn.Incoming = make(chan string)

	origin := "http://localhost/"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		fmt.Println("Error connecting to remote MosesACS instance")
		line.Close()
		os.Exit(1)
	}

	conn.ws = ws
	go conn.read()
}

func (conn *Connection) read() {
	var msg = make([]byte, 512)
	for {
		if _, err := conn.ws.Read(msg); err != nil {
			conn.Incoming <- "quit"
		}

		conn.Incoming <- string(msg)
	}
}

func (conn *Connection) Close() {
	conn.ws.Close()
}

func (conn *Connection) Write(cmd string) {
	var ws = conn.ws
	if _, err := ws.Write([]byte(cmd)); err != nil {
		conn.Incoming <- "quit"
	}
}
