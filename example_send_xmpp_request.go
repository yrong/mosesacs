package main

import (
  "./xmpp"
  "fmt"
)

func main() {
  xmpp.StartClient()
  defer xmpp.Close()
  xmpp.SendConnectionRequest("cpe@mosesacs.org")
}
