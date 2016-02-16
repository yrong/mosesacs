package daemon

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/lucacervasio/mosesacs/cwmp"
	"github.com/lucacervasio/mosesacs/www"
	"github.com/oleiade/lane"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const Version = "0.2.0"

var logger MosesWriter

type MosesWriter interface {
	Logger(string)
}

type BasicWriter struct {
}

func (w *BasicWriter) Logger(log string) {
	fmt.Println("Free:", log)
}

type Request struct {
	Id          string
	Websocket   *websocket.Conn
	CwmpMessage string
	Callback    func(msg *WsSendMessage) error
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
	DataModel            string
	KeepConnectionOpen   bool
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

func CwmpHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Server", "MosesACS "+Version)

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
				DataModel:            Inform.GetDataModelType(),
				KeepConnectionOpen:   false}
		}
		obj := cpes[Inform.DeviceId.SerialNumber]
		cpe := &obj
		cpe.LastConnection = time.Now().UTC()

		log.Printf("Received an Inform from %s (%d bytes) with SerialNumber %s and EventCodes %s", addr, len, Inform.DeviceId.SerialNumber, Inform.GetEvents())
		log.Printf("Soap envelope has mustUnderstand %s\n", envelope.Header.Id)
		logger.Logger("ciao")
		sendAll(fmt.Sprintf("Received an Inform from %s (%d bytes) with SerialNumber %s and EventCodes %s", addr, len, Inform.DeviceId.SerialNumber, Inform.GetEvents()))

		expiration := time.Now().AddDate(0, 0, 1) // expires in 1 day
		hash := "asdadasd"

		cookie := http.Cookie{Name: "mosesacs", Value: hash, Expires: expiration}
		http.SetCookie(w, &cookie)
		sessions[hash] = cpe

		fmt.Fprintf(w, cwmp.InformResponse(envelope.Header.Id))
	} else if messageType == "TransferComplete" {

	} else if messageType == "GetRPC" {

	} else {
		//		if messageType == "GetParameterValuesResponse" {
		// eseguo del parsing, invio i dati via websocket o altro

		//	} else if len == 0 {
		if len == 0 {
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

				cpe.Waiting.Callback(msg)
				//				if err := websocket.JSON.Send(cpe.Waiting.Websocket, msg); err != nil {
				//					fmt.Println("error while sending back answer:", err)
				//				}

			} else if e.KindOf() == "GetParameterValuesResponse" {
				var envelope cwmp.GetParameterValuesResponse
				xml.Unmarshal([]byte(body), &envelope)

				msg := new(WsSendMessage)
				msg.MsgType = "GetParameterValuesResponse"
				msg.Data, _ = json.Marshal(envelope)

				cpe.Waiting.Callback(msg)
				//				if err := websocket.JSON.Send(cpe.Waiting.Websocket, msg); err != nil {
				//					fmt.Println("error while sending back answer:", err)
				//				}

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
			if cpe.KeepConnectionOpen {
				fmt.Println("I'm keeping connection open")
			} else {
				w.WriteHeader(204)
			}
		}
	}

}

func doConnectionRequest(SerialNumber string) {
	fmt.Println("issuing a connection request to CPE", SerialNumber)
	//	http.Get(cpes[SerialNumber].ConnectionRequestURL)
	Auth("user", "pass", cpes[SerialNumber].ConnectionRequestURL)
}

func staticPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, www.Index)
}

func fontsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./www"+r.URL.Path)
}

func Run(port *int, logObj MosesWriter) {
	logger = logObj
	cpes = make(map[string]CPE)
	sessions = make(map[string]*CPE)

	// plain http handler for cpes
	fmt.Printf("HTTP Handler installed at http://0.0.0.0:%d/acs for cpes to connect\n", *port)
	http.HandleFunc("/acs", CwmpHandler)

	fmt.Printf("Websocket API endpoint installed at http://0.0.0.0:%d/api for admin stuff\n", *port)
	http.Handle("/api", websocket.Handler(websocketHandler))

	fmt.Printf("WEB Handler installed at http://0.0.0.0:%d/www\n", *port)
	http.HandleFunc("/www", staticPage)
	http.HandleFunc("/fonts/", fontsPage)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
