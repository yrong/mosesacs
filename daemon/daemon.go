package daemon

import (
	"code.google.com/p/go.net/websocket"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	//"encoding/json"
	"github.com/lucacervasio/mosesacs/cwmp"
	"github.com/lucacervasio/mosesacs/www"
	"github.com/oleiade/lane"
	"strings"
	"time"
	//	"regexp"
	//	"strconv"
	"encoding/json"
)

const Version = "0.1.10"

var logger MosesWriter

type MosesWriter interface {
	Logger(string)
}

type BasicWriter struct {

}

func (w *BasicWriter) Logger(log string) {
	fmt.Println("Free:",log)
}

type Request struct {
	Id          string
	Websocket   *websocket.Conn
	CwmpMessage string
}

type CPE struct {
	SerialNumber         string
	Manufacturer         string
	OUI                  string
	ConnectionRequestURL string
	SoftwareVersion      string
	ExternalIPAddress    string
	State                string
	Queue                *lane.Queue
	Waiting              *Request
	HardwareVersion      string
	LastConnection       time.Time
	DataModel			 string
}

type Message struct {
	SerialNumber string
	Message      string
}

type WsMessage struct {
	Cmd string
}

type WsSendMessage struct {
	MsgType string
	Data    json.RawMessage
}

type MsgCPEs struct {
	CPES map[string]CPE
}

var cpes map[string]CPE      // by serial
var sessions map[string]*CPE // by session cookie
var clients []Client

//func (req Request) reply(msg string) {
//	if _, err := req.Websocket.Write([]byte(msg)); err != nil {
//
//	}
//}

func handler(w http.ResponseWriter, r *http.Request) {
	//	log.Printf("New connection coming from %s", r.RemoteAddr)
	defer r.Body.Close()
	tmp, _ := ioutil.ReadAll(r.Body)

	body := string(tmp)
	len := len(body)

	//	log.Printf("body: %v", body)
	//	log.Printf("body length: %v", len)

	var envelope cwmp.SoapEnvelope
	xml.Unmarshal(tmp, &envelope)

	messageType := envelope.Body.CWMPMessage.XMLName.Local

	var cpe *CPE

	if messageType != "Inform" {
		if cookie, err := r.Cookie("mosesacs"); err == nil {
			cpe = sessions[cookie.Value]
		} else {
			fmt.Println("cookie 'mosesacs' missing")
			w.WriteHeader(401)
			return
		}
	}

	if messageType == "Inform" {
		var Inform cwmp.CWMPInform
		xml.Unmarshal(tmp, &Inform)

		var addr string
		if r.Header.Get("X-Real-Ip") != "" {
			addr = r.Header.Get("X-Real-Ip")
		} else {
			addr = r.RemoteAddr
		}

		if _, exists := cpes[Inform.DeviceId.SerialNumber]; !exists {
			fmt.Println("found ConnectionRequest " + Inform.GetConnectionRequest())
			cpes[Inform.DeviceId.SerialNumber] = CPE{
				SerialNumber:         Inform.DeviceId.SerialNumber,
				LastConnection:       time.Now().UTC(),
				SoftwareVersion:      Inform.GetSoftwareVersion(),
				HardwareVersion:      Inform.GetHardwareVersion(),
				ExternalIPAddress:    addr,
				ConnectionRequestURL: Inform.GetConnectionRequest(),
				OUI:                  Inform.DeviceId.OUI,
				Queue:                lane.NewQueue(),
				DataModel:			  Inform.GetDataModelType()}
		}
		obj := cpes[Inform.DeviceId.SerialNumber]
		cpe := &obj
		cpe.LastConnection = time.Now().UTC()

		log.Printf("Received an Inform from %s (%d bytes) with SerialNumber %s and EventCodes %s", addr, len, Inform.DeviceId.SerialNumber, Inform.GetEvents())
		logger.Logger("ciao")
		sendAll(fmt.Sprintf("Received an Inform from %s (%d bytes) with SerialNumber %s and EventCodes %s", addr, len, Inform.DeviceId.SerialNumber, Inform.GetEvents()))

		expiration := time.Now().AddDate(0, 0, 1) // expires in 1 day
		hash := "asdadasd"

		cookie := http.Cookie{Name: "mosesacs", Value: hash, Expires: expiration}
		http.SetCookie(w, &cookie)
		sessions[hash] = cpe

		fmt.Fprintf(w, cwmp.InformResponse())
	} else if messageType == "TransferComplete" {

	} else if messageType == "GetRPC" {

	} else {
		if messageType == "GetParameterValuesResponse" {
			// eseguo del parsing, invio i dati via websocket o altro
		} else if len == 0 {
			// empty post
			log.Printf("Got Empty Post")
		}

		if cpe.Waiting != nil {

			var e cwmp.SoapEnvelope
			xml.Unmarshal([]byte(body), &e)

			if e.KindOf() == "GetParameterNamesResponse" {
				var envelope cwmp.GetParameterNamesResponse
				xml.Unmarshal([]byte(body), &envelope)

				msg := new(WsSendMessage)
				msg.MsgType = "GetParameterNamesResponse"
				msg.Data, _ = json.Marshal(envelope)

				if err := websocket.JSON.Send(cpe.Waiting.Websocket, msg); err != nil {
					fmt.Println("error while sending back answer:", err)
				}

			} else {
				msg := new(WsMessage)
				msg.Cmd = body

				if err := websocket.JSON.Send(cpe.Waiting.Websocket, msg); err != nil {
					fmt.Println("error while sending back answer:", err)
				}

			}

			cpe.Waiting = nil
		}

		// Got Empty Post or a Response. Now check for any event to send, otherwise 204
		if cpe.Queue.Size() > 0 {
			req := cpe.Queue.Dequeue().(Request)
			// fmt.Println("sending "+req.CwmpMessage)
			fmt.Fprintf(w, req.CwmpMessage)
			cpe.Waiting = &req
		} else {
			w.WriteHeader(204)
		}
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

func websocketHandler(ws *websocket.Conn) {
	fmt.Println("New websocket client via ws")
	defer ws.Close()

	client := Client{ws: ws, start: time.Now().UTC()}
	clients = append(clients, client)
	//	client.Read()

	quit := make(chan bool)
	go periodicWsChecker(&client, quit)

	for {
		var msg WsMessage
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			fmt.Println("error while Receive:", err)
			quit <- true
			break
		}

		m := msg.Cmd
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
			req := Request{i[1], ws, cwmp.GetParameterValues(i[2])}

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
			req := Request{i[1], ws, cwmp.GetParameterNames(i[2])}

			if _, exists := cpes[i[1]]; exists {
				cpes[i[1]].Queue.Enqueue(req)
				if cpes[i[1]].State != "Connected" {
					// issue a connection request
					go doConnectionRequest(i[1])
				}
			} else {
				fmt.Println(fmt.Sprintf("CPE with serial %s not found", i[1]))
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

func doConnectionRequest(SerialNumber string) {
	fmt.Println("issuing a connection request to CPE", SerialNumber)
	//	http.Get(cpes[SerialNumber].ConnectionRequestURL)
	Auth("user", "pass", cpes[SerialNumber].ConnectionRequestURL)
}

func handlerWWW(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, www.Index)
}

func Run(port *int, logObj MosesWriter) {
	logger = logObj
	cpes = make(map[string]CPE)
	sessions = make(map[string]*CPE)

	// plain http handler for cpes
	fmt.Printf("HTTP Handler installed at http://0.0.0.0:%d/acs for cpes to connect\n", *port)
	http.HandleFunc("/acs", handler)

	fmt.Printf("Endpoint installed at http://0.0.0.0:%d/api for admin stuff\n", *port)
	http.Handle("/api", websocket.Handler(websocketHandler))

	fmt.Printf("WEB handler installed at http://0.0.0.0:%d/www\n", *port)
	http.HandleFunc("/www", handlerWWW)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
