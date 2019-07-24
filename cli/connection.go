package client

import (
	"fmt"
	"golang.org/x/net/websocket"
	"os"
	//	"strings"
	"encoding/json"
	"github.com/yrong/mosesacs/daemon"
)

type Connection struct {
	ws       *websocket.Conn
	Status   string
	url      string
	Incoming chan daemon.WsSendMessage
}

func (conn *Connection) Start(url string) {
	conn.url = url
	conn.Incoming = make(chan daemon.WsSendMessage)

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
		var msg daemon.WsSendMessage
		err := websocket.JSON.Receive(conn.ws, &msg)
		if err != nil {
			fmt.Println("error while Reading:", err)
			//			conn.Incoming <- "quit"
			break
		}

		if msg.MsgType == "ping" {
			conn.Write("pong")
		} else {
			conn.Incoming <- msg
		}
	}
}

func (conn *Connection) Close() {
	conn.ws.Close()
}

func (conn *Connection) Write(cmd string) {
	msg := new(daemon.WsSendMessage)
	msg.MsgType = "command"

	var temp = make(map[string]string)
	temp["command"] = cmd
	msg.Data, _ = json.Marshal(temp)

	err := websocket.JSON.Send(conn.ws, msg)
	if err != nil {
		fmt.Println("error while Writing:", err)
		//		conn.Incoming <- "quit"
	}
}

func (conn *Connection) SendSyncCommand(cmd string) *daemon.WsSendMessage {
	ch := make(chan *daemon.WsSendMessage)
	conn.Write(cmd)
	m := <-ch

	return m
}
