package main

import (
	"github.com/lucacervasio/mosesacs/client"
)

func main() {
	agent := cwmpclient.NewClient()
	agent.Run()
}

