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
)

type SoapEnvelope struct {
	XMLName xml.Name
	Header SoapHeader
	Body    SoapBody
}
type SoapHeader struct {}
type SoapBody struct {
	CWMPMessage		   CWMPMessage    `xml:",any"`
}
type CWMPMessage struct {
	XMLName xml.Name

}


type CWMPInform struct {
	DeviceId	DeviceID   `xml:"Body>Inform>DeviceId"`
	Events		[]Event
}
type DeviceID struct {
	Manufacturer string
	OUI	string
	SerialNumber string
}

type Event struct {

}


type Message struct {
	SerialNumber string
	Message	string
}

type CPE struct {
	SerialNumber string
	Manufacturer string
	OUI string
	ConnectionRequestURL string
	SoftwareVersion string
	ExternalIPAddress string
	State string
}

var cpes  map[string]CPE

func informResponse() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:schemaLocation="urn:dslforum-org:cwmp-1-0 ..\schemas\wt121.xsd" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soap:Header/>
  <soap:Body soap:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
    <cwmp:InformResponse>
      <MaxEnvelopes>1</MaxEnvelopes>
    </cwmp:InformResponse>
  </soap:Body>
</soap:Envelope>`
}

func handler(w http.ResponseWriter, r *http.Request) {
//	log.Printf("New connection coming from %s", r.RemoteAddr)
	defer r.Body.Close()
	tmp, _ := ioutil.ReadAll(r.Body)

	body := string(tmp)
	len := len(body)

//	log.Printf("body: %v", body)
//	log.Printf("body length: %v", len)

	var envelope SoapEnvelope
	xml.Unmarshal(tmp, &envelope)

	messageType := envelope.Body.CWMPMessage.XMLName.Local


	if messageType == "Inform" {
		var Inform CWMPInform
		xml.Unmarshal(tmp, &Inform)
		fmt.Println(Inform)

		log.Printf("Received an Inform from %s (%d bytes)", r.RemoteAddr, len)

		fmt.Fprintf(w, informResponse())
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

		m, err := ws.Write(msg[:n])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Send: %s\n", msg[:m])
	}
}

func EchoServer(ws *websocket.Conn) {
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Receive: %s\n", msg[:n])
	ws.Write(msg[:n])
//	io.Copy(ws, ws)
}

func doConnectionRequest(SerialNumber string) {
	http.Get(cpes[SerialNumber].ConnectionRequestURL)
}

func main() {
	cpes = make(map[string]CPE)

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





