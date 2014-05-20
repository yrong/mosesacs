package client

import (
        "fmt"
        "log"
        "code.google.com/p/go.net/websocket"
)

// This example demonstrates a trivial client.
func Connect(url string) {
        origin := "http://localhost/"
        ws, err := websocket.Dial(url, "", origin)
        if err != nil {
                log.Fatal(err)
        }
        if _, err := ws.Write([]byte("hello, world!\n")); err != nil {
                log.Fatal(err)
        }
        log.Println("sent")
        var msg = make([]byte, 512)
        var n int
        fmt.Println("waiting to read...")
        for {
          if n, err = ws.Read(msg); err != nil {
                  log.Fatal(err)
          }
          fmt.Printf("Received: %s.\n", msg[:n])
        }
}
