package client

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"os"
	//	"strings"
	"github.com/lucacervasio/mosesacs/daemon"
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
	for {
		var msg daemon.WsMessage
		err := websocket.JSON.Receive(conn.ws, &msg)
		if err != nil {
			fmt.Println("error while Reading:",err)
			conn.Incoming <- "quit"
			break
		}

		if msg.Cmd == "ping" {
			conn.Write("pong")
		} else {
			conn.Incoming <- msg.Cmd
		}
	}
}

func (conn *Connection) Close() {
	conn.ws.Close()
}

func (conn *Connection) Write(cmd string) {
	msg := new(daemon.WsMessage)
	msg.Cmd = cmd

	err := websocket.JSON.Send(conn.ws, msg)
	if err != nil {
		fmt.Println("error while Writing:",err)
		conn.Incoming <- "quit"
	}
}
