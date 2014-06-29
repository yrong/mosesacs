package daemon

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//	"strings"
	"code.google.com/p/go.net/websocket"
	"encoding/xml"
	"os"
	//"encoding/json"
	"github.com/lucacervasio/mosesacs/cwmp"
	"github.com/oleiade/lane"
	"strings"
	"time"
	"regexp"
	"strconv"
)

const Version = "0.1.1"

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
	Queue 				 *lane.Queue
	Waiting				 *Request
}

type Message struct {
	SerialNumber string
	Message      string
}

var cpes map[string]CPE      // by serial
var sessions map[string]*CPE  // by session cookie
var clients []Client


func (req Request) reply(msg string) {
	if _, err := req.Websocket.Write([]byte(msg)); err != nil {

	}
}

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
			fmt.Println("recuperato CPE",cpe)
		} else {
			fmt.Println("cookie 'mosesacs' missing")
			w.WriteHeader(401)
			return
		}
	}

	if messageType == "Inform" {
		var Inform cwmp.CWMPInform
		xml.Unmarshal(tmp, &Inform)
		fmt.Println(Inform)

		fmt.Println("Serial:", Inform.DeviceId.SerialNumber)

		if _,exists := cpes[Inform.DeviceId.SerialNumber]; !exists {
			cpes[Inform.DeviceId.SerialNumber] = CPE{SerialNumber: Inform.DeviceId.SerialNumber, OUI: Inform.DeviceId.OUI, Queue: lane.NewQueue()}
		}
		obj := cpes[Inform.DeviceId.SerialNumber]
		cpe := &obj

		log.Printf("Received an Inform from %s (%d bytes)", r.RemoteAddr, len)

		expiration := time.Now().AddDate(0,0,1)
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
			fmt.Println("devo mandare indietro questa risposta")
			if _, err := cpe.Waiting.Websocket.Write([]byte(body)); err != nil {
				fmt.Println(err)
			}
			cpe.Waiting = &Request{}
		}

		// Got Empty Post or a Response. Now check for any event to send, otherwise 204
		if cpe.Queue.Size() > 0 {
			req := cpe.Queue.Dequeue().(Request)
//			fmt.Println("sending "+req.CwmpMessage)
			fmt.Fprintf(w, req.CwmpMessage)
			cpe.Waiting = &req
			fmt.Println("setting Waiting to req",req)
			fmt.Println("cpe vale",cpe)
		} else {
			w.WriteHeader(204)
		}
	}

}

func websocketHandler(ws *websocket.Conn) {
	fmt.Println("New websocket client via ws")
	defer ws.Close()

	client := Client{ws: ws, start: time.Now().UTC()}
	clients = append(clients, client)
//	client.Read()

	msg := make([]byte, 512)

	for {
		n := 0
		n, err := ws.Read(msg)
		//			fmt.Printf("Letti %d bytes \n",n)
		if err != nil {
			fmt.Println("Error while reading from remote websocket")
			break
		}
		// fmt.Printf("R: <%s>\n",msg[:n])
		m := strings.Trim(string(msg[:n]), "\r\n"+string(0))
		//			fmt.Printf("Received: <%s>\n", m)

		r, _ := regexp.Compile("readMib")
		// matched, err := regexp.MatchString("readMib", m)
		// fmt.Println(matched, err)

		if m == "list" {
			//				fmt.Println("cpes list")
			var cpeListMessage string

			for key, value := range cpes {
				fmt.Println("Key:", key, "Value:", value.OUI)
//				cpeListMessage += "CPE #" + key + " with OUI " + value.OUI + " ["+value.Queue.Size()+"]\n"
				cpeListMessage += fmt.Sprintf("CPE #%s with OUI %s [%d]\n", key, value.OUI, value.Queue.Size())
				// strings.Join(cpeListMessage, "CPE #"+key+" with OUI "+value.OUI+"\n")
			}

			_, err := ws.Write([]byte(cpeListMessage))
			if err != nil {
				fmt.Println("Error while writing to remote websocket")
				break
			}

			// client requests a GetParametersValues to cpe with serial
			//serial := "1"
			//leaf := "Device.Time."
			// enqueue this command with the ws number to get the answer back

		} else if m == "version" {
			_, err := ws.Write([]byte(fmt.Sprintf("MosesAcs Daemon %s", Version)))
			if err != nil {
				fmt.Println("Error while writing to remote websocket")
				break
			}

		} else if m == "status" {
			var response string
			for i:= range clients {
				response += clients[i].String()
			}

			_, err := ws.Write([]byte(response))
			if err != nil {
				fmt.Println("Error while writing to remote websocket")
				break
			}
		} else if r.MatchString(m) == true {
			fmt.Println("READ MIB")
			re := regexp.MustCompile(`\s`)
			i := re.Split("readMib 2 InternetGatewayDevice.", -1)
			cpeSerial, _ := strconv.Atoi(i[1])
			fmt.Printf("CPE %d\n", cpeSerial)
			fmt.Printf("LEAF %s\n", i[2])
			req := Request{"1", ws, cwmp.GetParameterValues(i[2])}

			if _,exists := cpes[i[1]]; exists {
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
	fmt.Println("leaving from read routine")

}

func doConnectionRequest(SerialNumber string) {
	fmt.Println("issuing a connection request to CPE", SerialNumber)
	http.Get(cpes[SerialNumber].ConnectionRequestURL)
}

func Run(port *int) {
	cpes = make(map[string]CPE)
	sessions = make(map[string]*CPE)

	// plain http handler for cpes
	fmt.Printf("HTTP Handler installed at http://0.0.0.0:%d/acs for cpes to connect\n", *port)
	http.HandleFunc("/acs", handler)

	fmt.Printf("Endpoint installed at http://0.0.0.0:%d/api for admin stuff\n", *port)
	http.Handle("/api", websocket.Handler(websocketHandler))

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
