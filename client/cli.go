package client

import (
	"fmt"
	"github.com/lucacervasio/liner"
	"os"
	"os/signal"
	"strings"
//	"time"
)

var line *liner.State
var client Connection

func Run(url string) {
	line = liner.NewLiner()
	defer line.Close()

	client.Start(fmt.Sprintf("ws://%s/api", url))
	defer client.Close()

	fmt.Printf("Connected to MosesACS @ws://%s/api\n", url)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			if sig.String() == "interrupt" {
				fmt.Printf("\n")
				quit(url, line)
			}
		}
	}()

	baseCmds := []string{"exit", "help", "version", "list", "status", "shutdown", "uptime", "readMib", "GetParameterNames"}

	line.SetCompleter(func(line string) (c []string) {
		for _, n := range baseCmds {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	if f, err := os.Open(fmt.Sprintf("/Users/lc/.moses@%s.history", url)); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	go receiver()

	for {

		if cmd, err := line.Prompt(fmt.Sprintf("moses@%s> ", url)); err != nil {
			fmt.Println("Error reading line: ", err)
		} else {
			// add to history
			if cmd == "exit" {
				quit(url, line)
			} else if cmd != "" && cmd != "\n" && cmd != "\r\n" {
				line.AppendHistory(cmd)
				processCommand(cmd)
			}
		}

	}

	// quit
	quit(url, line)
}

func receiver() {
	for {
		msg := <-client.Incoming
		if msg == "quit" {
			quit("TODO", line)
		}
		line.PrintAbovePrompt(string(msg))
	}
}

func quit(url string, line *liner.State) {
	if f, err := os.Create(fmt.Sprintf("/Users/lc/.moses@%s.history", url)); err != nil {
		fmt.Println("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}

	line.Close()
	fmt.Println("Disconnected. Bye.")
	os.Exit(0)
}

func processCommand(cmd string) {
	switch {
	case strings.Contains(cmd, "version"):
		client.Write("version")
	case strings.Contains(cmd, "readMib"):
		client.Write(cmd)
	case strings.Contains(cmd, "GetParameterNames"):
		client.Write(cmd)
	case strings.Contains(cmd, "list"):
		client.Write("list")
	case strings.Contains(cmd, "status"):
		client.Write("status")
	default:
		fmt.Println("Unknown command")
	}
}
