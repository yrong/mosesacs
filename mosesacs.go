package main

import (
  "fmt"
  "github.com/lucacervasio/mosesacs/daemon"
  "github.com/lucacervasio/mosesacs/client"
)

func main() {
  fmt.Println("Running mosesacs daemon")
  if (true) {
    daemon.Run()
  } else {
    client.Connect("qui")
  }
}
