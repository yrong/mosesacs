package daemon

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"time"
	//	"github.com/lucacervasio/mosesacs/client"
	"encoding/json"
)

type Client struct {
	ws    *websocket.Conn
	start time.Time
}

func (client *Client) String() string {
	uptime := time.Now().UTC().Sub(client.start)
	var addr string
	if client.ws.Request().Header.Get("X-Real-Ip") != "" {
		addr = client.ws.Request().Header.Get("X-Real-Ip")
	} else {
		addr = client.ws.Request().RemoteAddr
	}
	return fmt.Sprintf("%s has been up for %s", addr, uptime)
}

func (client *Client) Send(cmd string) {
	msg := new(WsSendMessage)
	msg.MsgType = "log"

	m := make(map[string]string)
	m["log"] = cmd
	msg.Data, _ = json.Marshal(m)

	client.SendNew(msg)
}

func (client *Client) SendNew(msg *WsSendMessage) {
	err := websocket.JSON.Send(client.ws, msg)
	if err != nil {
		fmt.Println("error while Writing:", err)
	}
}
