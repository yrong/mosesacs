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
	"time"
)

const Version = "0.1.1"

type Message struct {
	SerialNumber string
	Message      string
}

var cpes map[string]cwmp.CPE

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

		fmt.Println("Serial:", Inform.DeviceId.SerialNumber)

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

func websocketHandler(ws *websocket.Conn) {
	fmt.Println("New websocket client via ws")
	defer ws.Close()

	msg := make([]byte, 512)

	go func() {
		for {
			n, err := ws.Read(msg)
			if err != nil {
				fmt.Println("Error while reading from remote websocket")
				break
			}
			fmt.Printf("Received: %s", msg[:n])
      if string(msg) == "list" {
        // client requests a GetParametersValues to cpe with serial
        //serial := "1"
        //leaf := "Device.Time."
        // enqueue this command with the ws number to get the answer back
      }
		}
		fmt.Println("leaving from read routine")
	}()

	for {
		_, err := ws.Write([]byte("ciao"))
		if err != nil {
			fmt.Println("Error while writing to remote websocket")
			break
		}
		fmt.Printf("Send: %s\n", "ciao")
		time.Sleep(2 * time.Second)
	}
	fmt.Println("leaving from write routine")

	fmt.Println("websocket client has gone")
}

func doConnectionRequest(SerialNumber string) {
	http.Get(cpes[SerialNumber].ConnectionRequestURL)
}

func Run(port *int) {
	cpes = make(map[string]cwmp.CPE)

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
