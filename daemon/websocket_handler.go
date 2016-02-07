package daemon

import (
	"golang.org/x/net/websocket"
	"encoding/json"
	"fmt"
	"github.com/lucacervasio/mosesacs/cwmp"
	"strings"
	"time"
)

func websocketHandler(ws *websocket.Conn) {
	fmt.Println("New websocket client via ws")
	defer ws.Close()

	client := Client{ws: ws, start: time.Now().UTC()}
	clients = append(clients, client)
	//	client.Read()

	quit := make(chan bool)
	go periodicWsChecker(&client, quit)

	for {
		var msg WsSendMessage
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			fmt.Println("error while Receive:", err)
			quit <- true
			break
		}

		data := make(map[string]string)
		err = json.Unmarshal(msg.Data, &data)

		if err != nil {
			fmt.Println("error:", err)
		}

		m := data["command"]
		fmt.Println(m)

		if m == "list" {

			ms := new(WsSendMessage)
			ms.MsgType = "cpes"
			msgCpes := new(MsgCPEs)
			msgCpes.CPES = cpes
			ms.Data, _ = json.Marshal(msgCpes)

			client.SendNew(ms)

			// client requests a GetParametersValues to cpe with serial
			//serial := "1"
			//leaf := "Device.Time."
			// enqueue this command with the ws number to get the answer back

		} else if m == "version" {
			client.Send(fmt.Sprintf("MosesAcs Daemon %s", Version))

		} else if m == "status" {
			var response string
			for i := range clients {
				response += clients[i].String() + "\n"
			}

			client.Send(response)

		} else if strings.Contains(m, "readMib") {
			i := strings.Split(m, " ")
			//			cpeSerial, _ := strconv.Atoi(i[1])
			//			fmt.Printf("CPE %d\n", cpeSerial)
			//			fmt.Printf("LEAF %s\n", i[2])
			req := Request{i[1], ws, cwmp.GetParameterValues(i[2]), func(msg *WsSendMessage) error {
				if err := websocket.JSON.Send(ws, msg); err != nil {
					fmt.Println("error while sending back answer:", err)
				}

				return err
			}}

			if _, exists := cpes[i[1]]; exists {
				cpes[i[1]].Queue.Enqueue(req)
				if cpes[i[1]].State != "Connected" {
					// issue a connection request
					go doConnectionRequest(i[1])
				}
			} else {
				fmt.Println(fmt.Sprintf("CPE with serial %s not found", i[1]))
			}

		} else if strings.Contains(m, "writeMib") {
			i := strings.Split(m, " ")
			req := Request{i[1], ws, cwmp.SetParameterValues(i[2], i[3]), func(msg *WsSendMessage) error {
				if err := websocket.JSON.Send(ws, msg); err != nil {
					fmt.Println("error while sending back answer:", err)
				}

				return err
			}}

			if _, exists := cpes[i[1]]; exists {
				cpes[i[1]].Queue.Enqueue(req)
				if cpes[i[1]].State != "Connected" {
					// issue a connection request
					go doConnectionRequest(i[1])
				}
			} else {
				fmt.Println(fmt.Sprintf("CPE with serial %s not found", i[1]))
			}
		} else if strings.Contains(m, "GetParameterNames") {
			i := strings.Split(m, " ")
			req := Request{i[1], ws, cwmp.GetParameterNames(i[2]), func(msg *WsSendMessage) error {
				if err := websocket.JSON.Send(ws, msg); err != nil {
					fmt.Println("error while sending back answer:", err)
				}

				return err
			}}

			if _, exists := cpes[i[1]]; exists {
				cpes[i[1]].Queue.Enqueue(req)
				if cpes[i[1]].State != "Connected" {
					// issue a connection request
					go doConnectionRequest(i[1])
				}
			} else {
				fmt.Println(fmt.Sprintf("CPE with serial %s not found", i[1]))
			}
		} else if m == "GetParameterValues" {
			cpe := data["cpe"]
			req := Request{cpe, ws, cwmp.GetParameterValues(data["object"]), func(msg *WsSendMessage) error {
				if err := websocket.JSON.Send(ws, msg); err != nil {
					fmt.Println("error while sending back answer:", err)
				}

				return err
			}}
			if _, exists := cpes[cpe]; exists {
				cpes[cpe].Queue.Enqueue(req)
				if cpes[cpe].State != "Connected" {
					// issue a connection request
					go doConnectionRequest(cpe)
				}
			} else {
				fmt.Println(fmt.Sprintf("CPE with serial %s not found", cpe))
			}
		} else if m == "GetSummary" {
			cpe := data["cpe"]
			ch := make(chan *WsSendMessage)
			req := Request{cpe, ws, cwmp.GetParameterValues(data["object"]), func(msg *WsSendMessage) error {
				fmt.Println("sono nella callback")
				ch <- msg
				return nil // TODO da implementare un timeout ? boh
			}}
			if _, exists := cpes[cpe]; exists {
				cpes[cpe].Queue.Enqueue(req)
				if cpes[cpe].State != "Connected" {
					// issue a connection request
					go doConnectionRequest(cpe)
				}
			} else {
				fmt.Println(fmt.Sprintf("CPE with serial %s not found", cpe))
			}

			fmt.Println("sono sospeso in attesa che ritorni il messaggio")
			m := <- ch
			fmt.Println("è tornato")
			if err := websocket.JSON.Send(ws, m); err != nil {
				fmt.Println("error while sending back answer:", err)
			}

		} else if m == "getMib" {
			cpe := data["cpe"]
			req := Request{cpe, ws, cwmp.GetParameterNames(data["object"]), func(msg *WsSendMessage) error {
				fmt.Println("sono nella callback")
				if err := websocket.JSON.Send(ws, msg); err != nil {
					fmt.Println("error while sending back answer:", err)
				}

				return err
			}}
			if _, exists := cpes[cpe]; exists {
				cpes[cpe].Queue.Enqueue(req)
				if cpes[cpe].State != "Connected" {
					// issue a connection request
					go doConnectionRequest(cpe)
				}
			} else {
				fmt.Println(fmt.Sprintf("CPE with serial %s not found", cpe))
			}
		}
	}
	fmt.Println("ws closed, leaving read routine")

	for i := range clients {
		if clients[i].ws == ws {
			clients = append(clients[:i], clients[i+1:]...)
		}
	}
}

func sendAll(msg string) {
	for i := range clients {
		clients[i].Send(msg)
	}
}

func periodicWsChecker(c *Client, quit chan bool) {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Println("new tick on client:", c)
			c.Send("ping")
		case <-quit:
			fmt.Println("received quit command for periodicWsChecker")
			ticker.Stop()
			return
		}
	}
}
