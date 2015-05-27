package main

import (
	"github.com/lucacervasio/mosesacs/client"
	"log"
	"time"
)

func main() {
	log.Println("start cpes")
	cwmpclient.RunClient(5)
	time.Sleep(1000000000000)
}

