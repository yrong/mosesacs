package main

import (
  "fmt"
  "flag"
  "github.com/lucacervasio/mosesacs/daemon"
  "github.com/lucacervasio/mosesacs/client"
)

func main() {

  flDaemon := flag.Bool("d", false, "Enable daemon mode")
  flag.Parse()

  fmt.Println("Running mosesacs daemon")
  if (*flDaemon) {
    daemon.Run()
  } else {
    client.ExampleDial()
  }
}
