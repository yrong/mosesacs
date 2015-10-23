package cwmpclient

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"bytes"
	"github.com/lucacervasio/mosesacs/cwmp"
	"github.com/lucacervasio/mosesacs/xmpp"
	"net/url"
	"encoding/xml"
	"io/ioutil"
	"time"
	"math/rand"
	"strconv"
	"github.com/lucacervasio/mosesacs/daemon"
	"fmt"
)

type Agent struct {
	Status string
	AcsUrl string
	Cpe daemon.CPE
}

func NewClient() (agent Agent) {
	serial := strconv.Itoa(random(1000,5000))
	connection_request_url := "/ConnectionRequest-"+serial
	cpe := daemon.CPE{serial, "MOONAR LABS", "001309", connection_request_url, "asd", "asd", "0 BOOTSTRAP", nil, &daemon.Request{}, "4324asd", time.Now().UTC(), "TR181", false}
	agent = Agent{"initializing", "http://localhost:9292/acs", cpe}
	log.Println(agent)
	return
}

func (a Agent) String() string {
	return fmt.Sprintf("Agent running with serial %s and connection request url %s\n", a.Cpe.SerialNumber, a.Cpe.ConnectionRequestURL)
}

func (a Agent) Run() {
	http.HandleFunc(a.Cpe.ConnectionRequestURL, a.connectionRequestHandler)
	log.Println("Start http server waiting connection request")
	a.startConnection()
//  a.startXmppConnection()

	http.ListenAndServe(":7547", nil)
}

func (a Agent) connectionRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got connection request, send Inform to %s", a.AcsUrl)
	a.startConnection()
}

func random(min, max int) int {
	rand.Seed(int64(time.Now().Nanosecond()))
	return rand.Intn(max - min) + min
}

func (a Agent) startXmppConnection() {
  log.Println("starting StartXmppConnection") 
  xmpp.StartClient("cpe2@mosesacs.org", "password1234", func(str string){
    log.Println("got "+str)
  })
}

func (a Agent) startConnection(){
	log.Printf("send Inform to %s", a.AcsUrl)
	var msgToSend []byte
	msgToSend = []byte(cwmp.Inform(a.Cpe.SerialNumber))

	tr := &http.Transport{}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Transport: tr, Jar:jar}
	envelope := cwmp.SoapEnvelope{}
	u, _ := url.Parse(a.AcsUrl)

	resp, err := client.Post(a.AcsUrl, "text/xml", bytes.NewBuffer(msgToSend))
	if err != nil {
		log.Fatal("server unavailable")
	}
	log.Println(resp.Header)
	for {
		if resp.ContentLength == 0 {
			log.Println("got empty post, close connection")
			resp.Body.Close()
			tr.CloseIdleConnections()
			break
		} else {
			tmp, _ := ioutil.ReadAll(resp.Body)
			body := string(tmp)
			xml.Unmarshal(tmp, &envelope)

			if envelope.KindOf() == "GetParameterValues" {
				log.Println("Send GetParameterValuesResponse")
				var leaves cwmp.GetParameterValues_
				xml.Unmarshal([]byte(body), &leaves)
				msgToSend = []byte(cwmp.BuildGetParameterValuesResponse(a.Cpe.SerialNumber, leaves))
			} else if envelope.KindOf() == "GetParameterNames" {
				log.Println("Send GetParameterNamesResponse")
				var leaves cwmp.GetParameterNames_
				xml.Unmarshal([]byte(body), &leaves)
				msgToSend = []byte(cwmp.BuildGetParameterNamesResponse(a.Cpe.SerialNumber, leaves))
			} else {
				log.Println("send empty post")
				msgToSend = []byte("")
			}

			client.Jar.SetCookies(u, resp.Cookies())
			resp, _ = client.Post(a.AcsUrl, "text/xml", bytes.NewBuffer(msgToSend))
			log.Println(resp.Header)
		}
	}

}

