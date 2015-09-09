package main

import (
  "github.com/lucacervasio/mosesacs/xmpp"
  "time"
  "log"
)

func main() {
  xmpp.StartClient("acs@mosesacs.org", "password1234", func(str string) {
    log.Println(str)
  })
  defer xmpp.Close()
  xmpp.SendConnectionRequest("cpe1@mosesacs.org/casatua")
  time.Sleep(2 * time.Second)
}
