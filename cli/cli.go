package client

import (
	"encoding/json"
	//	"encoding/xml"
	"fmt"
	"github.com/lucacervasio/liner"
	"github.com/yrong/mosesacs/cwmp"
	"github.com/yrong/mosesacs/daemon"
	"os"
	"os/signal"
	"strings"
)

var line *liner.State
var client Connection
var context string

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

	baseCmds := []string{"exit", "setxmpp", "help", "version", "list", "status", "shutdown", "uptime", "readMib", "writeMib", "GetParameterNames", "set"}
	contextCmds := []string{"summary"}

	line.SetCompleter(func(line string) (c []string) {
		if strings.HasPrefix(line, "cpe ") {
			// should return the list of cpes as second argument
			arr := strings.Split(line, " ")

			cpes := []string{"cpe1", "cpe12", "cpe3"} // TODO get cpe list via ws
			for _, n := range cpes {
				if strings.HasPrefix(n, strings.ToLower(arr[1])) {
					c = append(c, arr[0]+" "+n)
				}
			}
		} else {
			// cycle through all available commands
			var cmds []string
			if context != "" {
				// if in context cycle through cpe-specific commands
				cmds = contextCmds
			} else {
				// otherwise cycle though base commands
				cmds = baseCmds
			}

			for _, n := range cmds {
				if strings.HasPrefix(n, strings.ToLower(line)) {
					c = append(c, n)
				}
			}
		}
		return
	})

	if f, err := os.Open(fmt.Sprintf(os.ExpandEnv("$HOME")+"/.moses@%s.history", url)); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	go receiver()
	context = ""

	for {
		if cmd, err := line.Prompt(fmt.Sprintf("moses@%s/%s> ", url, context)); err != nil {
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
		if msg.MsgType == "quit" {
			quit("TODO", line)
		}

		switch msg.MsgType {
		case "cpes":
			cpes := new(daemon.MsgCPEs)
			err := json.Unmarshal(msg.Data, &cpes)
			if err != nil {
				fmt.Println("error:", err)
			}

			line.PrintAbovePrompt("cpe list")
			for key, value := range cpes.CPES {
				line.PrintAbovePrompt(fmt.Sprintf("CPE %s with OUI %s", key, value.OUI))
			}
		case "GetParameterNamesResponse":
			getParameterNames := new(cwmp.GetParameterNamesResponse)
			err := json.Unmarshal(msg.Data, &getParameterNames)
			if err != nil {
				fmt.Println("error:", err)
			}
			//			fmt.Println(getParameterNames.ParameterList)
			for idx := range getParameterNames.ParameterList {
				line.PrintAbovePrompt(fmt.Sprintf("%s : %s", getParameterNames.ParameterList[idx].Name, getParameterNames.ParameterList[idx].Writable))
			}
		case "GetParameterValuesResponse":
			getParameterValues := new(cwmp.GetParameterValuesResponse)
			err := json.Unmarshal(msg.Data, &getParameterValues)
			if err != nil {
				fmt.Println("error:", err)
			}
			for idx := range getParameterValues.ParameterList {
				line.PrintAbovePrompt(fmt.Sprintf("%s : %s", getParameterValues.ParameterList[idx].Name, getParameterValues.ParameterList[idx].Value))
			}
		case "SetParameterValuesResponse":
			line.PrintAbovePrompt(fmt.Sprintf("got SetParameterValuesResponse"))
		case "log":
			log := make(map[string]string)
			err := json.Unmarshal(msg.Data, &log)
			if err != nil {
				fmt.Println("error:", err)
			}

			//			fmt.Printf("%+v",log["prova"])
			if log["log"] == "ping" {
				// received ping from daemon
			} else {
				line.PrintAbovePrompt(fmt.Sprintf("%s", log["log"]))
			}

		}

		//		fmt.Println(msg)
		/*
			var e cwmp.SoapEnvelope
			xml.Unmarshal([]byte(msg), &e)

			if e.KindOf() == "GetParameterValuesResponse" {
				var envelope cwmp.GetParameterValuesResponse
				xml.Unmarshal([]byte(msg), &envelope)

				for idx := range envelope.ParameterList {
					line.PrintAbovePrompt(string(fmt.Sprintf("%s : %s", envelope.ParameterList[idx].Name, envelope.ParameterList[idx].Value)))
				}

			} else if e.KindOf() == "GetParameterNamesResponse" {
				line.PrintAbovePrompt(string(msg))

				var envelope cwmp.GetParameterNamesResponse
				xml.Unmarshal([]byte(msg), &envelope)

				for idx := range envelope.ParameterList {
					line.PrintAbovePrompt(string(fmt.Sprintf("%s : %s", envelope.ParameterList[idx].Name, envelope.ParameterList[idx].Writable)))
				}

			} else {
				line.PrintAbovePrompt(string(msg))

			}

		*/
		//		line.PrintAbovePrompt(msg.MsgType)
	}
}

func quit(url string, line *liner.State) {
	if f, err := os.Create(fmt.Sprintf(os.ExpandEnv("$HOME")+"/.moses@%s.history", url)); err != nil {
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
	case strings.HasPrefix(cmd, "cpe "):
		arr := strings.Split(cmd, " ")

		context = arr[1]
	case strings.Contains(cmd, "version"):
		client.Write("version")
	case strings.Contains(cmd, "readMib"):
		client.Write(cmd)
	case strings.Contains(cmd, "writeMib"):
		client.Write(cmd)
//	case strings.Contains(cmd, "changeDuState"):
//		client.Write(cmd)
	case strings.Contains(cmd, "GetParameterNames"):
		client.Write(cmd)
	case strings.Contains(cmd, "list"):
		client.Write("list")
	case strings.Contains(cmd, "status"):
		client.Write("status")
	case strings.Contains(cmd, "setxmpp"):
		client.Write(cmd)
	default:
		fmt.Println("Unknown command")
	}
}
