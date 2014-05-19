package main

import (
	"fmt"
	"net/http"
	"log"
	"io/ioutil"
//	"strings"
	"encoding/xml"
	"os"
	"code.google.com/p/go.net/websocket"
//	"io"
	"flag"
	"encoding/json"
  "github.com/lucacervasio/hercules/cwmp"
)



type Message struct {
	SerialNumber string
	Message	string
}

var cpes  map[string]cwmp.CPE



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


	if messageType == "Inform" {
		var Inform cwmp.CWMPInform
		xml.Unmarshal(tmp, &Inform)
		fmt.Println(Inform)

		fmt.Println("Serial:",Inform.DeviceId.SerialNumber)

		cpes[Inform.DeviceId.SerialNumber] = cwmp.CPE{SerialNumber: Inform.DeviceId.SerialNumber, OUI: Inform.DeviceId.OUI}

		log.Printf("Received an Inform from %s (%d bytes)", r.RemoteAddr, len)

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

		// Got Empty Post or a Response. Now check for any event to send, otherwise 204
		w.WriteHeader(204)
	}




}

func echoHandler(ws *websocket.Conn) {
	msg := make([]byte, 512)

	for {
		n, err := ws.Read(msg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Receive: %s\n", msg[:n])

//		m, err := ws.Write(msg[:n])
		txt,_ := json.Marshal(cpes)
		fmt.Println(string(txt))
		m, err := ws.Write(txt)
		fmt.Println(m)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Send: %s\n", txt)
	}
}

func EchoServer(ws *websocket.Conn) {
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received: %s\n", msg[:n])
	ws.Write(msg[:n])
//	io.Copy(ws, ws)
}

func doConnectionRequest(SerialNumber string) {
	http.Get(cpes[SerialNumber].ConnectionRequestURL)
}

func main() {
	cpes = make(map[string]cwmp.CPE)

	port := flag.Int("p", 9090, "Port to listen on")
	flag.Parse()

	http.HandleFunc("/acs", handler)
	http.Handle("/ws", websocket.Handler(echoHandler))
	fmt.Printf("Serving on %d\n",*port)
	err := http.ListenAndServe(fmt.Sprintf(":%d",*port), nil)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}





