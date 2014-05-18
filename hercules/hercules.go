package main

import (
	"fmt"
	"net/http"
	"bytes"
	"io"
	"io/ioutil"
//	"time"
	"os"
	"log"
  "time"
  "flag"
  "math/rand"
  "strconv"
)

var num_cpes = flag.Int("n", 2, "how many CPEs should I emulate ?")

var AcsUrl = "http://localhost:9292/acs"

type CPE struct {
	SerialNumber string
	Manufacturer string
	OUI string
	ConnectionRequestURL string
	SoftwareVersion string
	ExternalIPAddress string
	State string
}

func (cpe CPE) runConnection() {
//	fmt.Printf("[%s] connecting with state %s\n", cpe.SerialNumber, cpe.State)
	fmt.Printf("[%s] --> Starting connection to %s, sending Inform with eventCode %s\n", cpe.SerialNumber, AcsUrl, cpe.State)

	buf := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:cwmp="urn:dslforum-org:cwmp-1-0"
               xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <soap:Header/>
    <soap:Body soap:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
        <cwmp:Inform>
            <DeviceId>
                <Manufacturer>`+cpe.Manufacturer+`</Manufacturer>
                <OUI>`+cpe.OUI+`</OUI>
                <ProductClass>Router</ProductClass>
                <SerialNumber>`+cpe.SerialNumber+`</SerialNumber>
            </DeviceId>
            <Event>
                <EventStruct>
                    <EventCode>`+cpe.State+`</EventCode>
                    <CommandKey/>
                </EventStruct>
            </Event>
            <MaxEnvelopes>1</MaxEnvelopes>
            <CurrentTime>2003-01-01T05:36:55Z</CurrentTime>
            <RetryCount>0</RetryCount>
            <ParameterList soap-enc:arrayType="cwmp:ParameterValueStruct[7]">
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.HardwareVersion</Name>
                    <Value xsi:type="xsd:string">NGRG 2009</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.ProvisioningCode</Name>
                    <Value xsi:type="xsd:string">ABCD</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.SoftwareVersion</Name>
                    <Value xsi:type="xsd:string">`+cpe.SoftwareVersion+`</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.DeviceInfo.SpecVersion</Name>
                    <Value xsi:type="xsd:string">1.0</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.ManagementServer.ConnectionRequestURL</Name>
                    <Value xsi:type="xsd:string">http://10.19.0.`+cpe.SerialNumber+`:9600/`+cpe.SerialNumber+`</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.ManagementServer.ParameterKey</Name>
                    <Value xsi:type="xsd:string"/>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress
                    </Name>
                    <Value xsi:type="xsd:string">10.19.0.`+cpe.SerialNumber+`</Value>
                </ParameterValueStruct>
            </ParameterList>
        </cwmp:Inform>
    </soap:Body>
</soap:Envelope>`



  tr := &http.Transport{}
  client := &http.Client{ Transport: tr }

	resp, err := client.Post(AcsUrl, "text/xml", bytes.NewBufferString(buf))
	if err != nil {
		fmt.Println(fmt.Sprintf("Couldn't connect to %s",AcsUrl))
    os.Exit(1)
	}

	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	fmt.Printf("[%s] <-- ACS replied with statusCode: %d, content-lenght: %d\n", cpe.SerialNumber, resp.StatusCode, resp.ContentLength)
	//	tmp,_ := ioutil.ReadAll(resp.Body)
	//	fmt.Printf("body: %s", string(tmp))

	fmt.Printf("[%s] --> Sending empty POST\n", cpe.SerialNumber)
	resp, err = client.Post(AcsUrl, "text/xml", bytes.NewBufferString(""))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("[%s] <-- ACS replied with statusCode: %d, content-lenght: %d\n", cpe.SerialNumber, resp.StatusCode, resp.ContentLength)

  resp.Body.Close()

  tr.CloseIdleConnections()
}

func handler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("new connection Request")
}

func (cpe CPE) periodic(interval int) {
  fmt.Printf("Bootstrapping CPE #%s with interval %ds\n", cpe.SerialNumber, interval)
  cpe.runConnection()
  for {
    time.Sleep(time.Duration(interval) * time.Second)
    cpe.runConnection()
  }
}

func random(min, max int) int {
    rand.Seed(time.Now().UnixNano())
    return rand.Intn(max - min) + min
}

func main() {
  // create cpe struct

  flag.Parse()
  fmt.Println("Starting Hercules with",*num_cpes,"cpes")

  CPEs := []CPE{}

	// initialize CPEs and send bootstrap
	//cpe1 := CPE{"1", "PIRELLI BROADBAND SOLUTIONS", "0013C8", "asd", "asd", "asd", "0 BOOTSTRAP"}
//	cpe2 := CPE{"2", "Telsey", "0014", "asd", "asd", "asd", "1 BOOT"}

  for i:=1; i <= *num_cpes; i++ {
    tmp_cpe := CPE{strconv.Itoa(i), "PIRELLI BROADBAND SOLUTIONS", "0013C8", "asd", "asd", "asd", "0 BOOTSTRAP"}
	  CPEs = append(CPEs, tmp_cpe)
  }


//	fmt.Println(CPEs)

	for _, c := range(CPEs) {
		go c.periodic(random(10,120))
	}


	// TODO run httpserver to wait for connection
//	time.Sleep (3 * time.Second)
	http.HandleFunc("/acs", handler)
	fmt.Println("Listening connection request port on 9600")
	err := http.ListenAndServe(":9600", nil)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

}

