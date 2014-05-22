package client

import (
  "fmt"
  "strings"
  "github.com/GeertJohan/go.linenoise"
)

func RunCli(url string) {
  fmt.Printf("Connected to MosesACS @ws://%s/api\n", url)

  baseCmds := []string{"exit", "help", "version", "list", "status", "shutdown", "uptime"}

  completionHandler := func(in string) []string {
    out := []string{}
    for s := range baseCmds {
      if strings.HasPrefix(baseCmds[s], in) {
        out = append(out, baseCmds[s])
      }
    }

    return out
  }

  linenoise.SetCompletionHandler(completionHandler)
  linenoise.LoadHistory(fmt.Sprintf("/Users/lc/.moses@%s.history", url))

  for {
    cmd, err := linenoise.Line(fmt.Sprintf("moses@%s> ",url))
    if cmd == "exit" || err == linenoise.KillSignalError {
      break
    }

    // add to history
    if cmd != "" && cmd != "\n" {
      fmt.Println("Got", cmd)
      linenoise.AddHistory(cmd)
    }
  }

  // quit
  linenoise.SaveHistory(fmt.Sprintf("/Users/lc/.moses@%s.history",url))
  fmt.Println("Disconnected. Bye.")
}
