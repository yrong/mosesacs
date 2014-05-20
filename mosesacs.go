package main

import (
  "flag"
  "fmt"
  "github.com/lucacervasio/mosesacs/daemon"
  "github.com/lucacervasio/mosesacs/client"
)

func main() {

	port := flag.Int("p", 9292, "Port to listen on")
  flDaemon := flag.Bool("d", false, "Enable daemon mode")
  flag.Parse()

  fmt.Printf("MosesACS %s by Luca Cervasio <luca.cervasio@gmail.com> (C)2014 http://mosesacs.org\n", daemon.Version)

  if (*flDaemon) {
    daemon.Run(port)
  } else {
    client.Connect("ws://localhost:9292/api")
  }
}
