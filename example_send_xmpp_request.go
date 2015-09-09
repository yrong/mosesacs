package main

import (
  "./xmpp"
  "time"
)

func main() {
  xmpp.StartClient()
  defer xmpp.Close()
  xmpp.SendConnectionRequest("cpe1@mosesacs.org")
  time.Sleep(2 * time.Second)
}
