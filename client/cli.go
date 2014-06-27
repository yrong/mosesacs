package client

import (
	"fmt"
	"github.com/lucacervasio/liner"
	"os"
	"os/signal"
	"strings"
	"time"
)

var line *liner.State

func Run(url string) {
	go logReceiver()
	runCli(url)
}

func logReceiver() {
	for {
		line.PrintAbovePrompt("log")
		time.Sleep(2 * time.Second)
	}
}

func runCli(url string) {
	line = liner.NewLiner()
	defer line.Close()

	chan_request := make(chan string)
	go Connect(fmt.Sprintf("ws://%s/api", url), chan_request)
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

	baseCmds := []string{"exit", "help", "version", "list", "status", "shutdown", "uptime"}

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

	for {

		if cmd, err := line.Prompt(fmt.Sprintf("moses@%s> ", url)); err != nil {
			fmt.Println("Error reading line: ", err)
		} else {
			// add to history
			if cmd == "exit" {
				quit(url, line)
			} else if cmd != "" && cmd != "\n" && cmd != "\r\n" {
				line.AppendHistory(cmd)
				processCommand(cmd, chan_request)
			}
		}

	}

	// quit
	quit(url, line)
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

func processCommand(cmd string, c chan string) {
	switch cmd {
	case "version":
		fmt.Println("0.1.2")
	case "readmib":
		c <- "readMib 10 InternetGatewayDevice."
	case "list":
		c <- "list"	
	case "num":
		c <- "num"
	default:
		fmt.Println("Unknown command")
	}
}
