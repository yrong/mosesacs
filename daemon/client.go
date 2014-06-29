package daemon

import (
	"fmt"
	"code.google.com/p/go.net/websocket"
	"time"
)

type Client struct {
	ws *websocket.Conn
	start time.Time
}


func (client *Client) String() string {
	uptime := time.Now().UTC().Sub(client.start)
	return fmt.Sprintf("%s is up from %s", client.ws.Request().RemoteAddr, uptime)
}
