package client

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"os"
	//	"strings"
)

func Connect(url string, chan_request chan string) {
	origin := "http://localhost/"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		fmt.Println("Error connecting to remote MosesACS instance")
		line.Close()
		os.Exit(1)
	}

	var channel = make(chan string)
	go Write(ws, channel, "bella yo")
	go Read(ws, channel)

	for {
		select {
		case cmd := <-channel:
			fmt.Printf("Got '%s' from channel\n", cmd)

			switch {
			case cmd == "quit":
				fmt.Println("Quit")
				line.Close()
				os.Exit(0)
			}
		case request := <-chan_request:
			fmt.Printf("cli request %s", request)
			go Write(ws, channel, request)
		}

	}

}

func Write(ws *websocket.Conn, channel chan string, cmd string) {
	fmt.Println("<"+cmd+">")
	if _, err := ws.Write([]byte(cmd)); err != nil {
		channel <- "quit"
	}
}

func Read(ws *websocket.Conn, channel chan string) {
	var msg = make([]byte, 512)
	for {
		if _, err := ws.Read(msg); err != nil {
			channel <- "quit"
		}
		channel <- string(msg)
	}

}
