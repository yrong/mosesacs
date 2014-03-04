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
)


var AcsUrl = "http://localhost:9090/acs"

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
                    <Value xsi:type="xsd:string">http://10.19.0.2:9600/1</Value>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.ManagementServer.ParameterKey</Name>
                    <Value xsi:type="xsd:string"/>
                </ParameterValueStruct>
                <ParameterValueStruct xsi:type="cwmp:ParameterValueStruct">
                    <Name>InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress
                    </Name>
                    <Value xsi:type="xsd:string">10.19.0.2</Value>
                </ParameterValueStruct>
            </ParameterList>
        </cwmp:Inform>
    </soap:Body>
</soap:Envelope>`

	//	tr := &http.Transport{
	//		DisableKeepAlives: false,
	//	}
	//	client := &http.Client{Transport: tr}

	resp, err := http.Post(AcsUrl, "text/xml", bytes.NewBufferString(buf))
	if err != nil {
		fmt.Println(err)
	}


	//	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(buf))
	//	req.Header.Set("Content-Type", "application/json")
	//	req.Header.Set("Connection", "Keep-Alive")
	//	resp, _ := client.Do(req)

	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	//	resp.Close = false

	fmt.Printf("[%s] <-- ACS replied with statusCode: %d, content-lenght: %d\n", cpe.SerialNumber, resp.StatusCode, resp.ContentLength)
	//	tmp,_ := ioutil.ReadAll(resp.Body)
	//	fmt.Printf("body: %s", string(tmp))

	fmt.Printf("[%s] --> Sending empty POST\n", cpe.SerialNumber)
	resp, err = http.Post(AcsUrl, "text/xml", bytes.NewBufferString(""))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("[%s] <-- ACS replied with statusCode: %d, content-lenght: %d\n", cpe.SerialNumber, resp.StatusCode, resp.ContentLength)

	//	resp.Close = true
}

func handler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// create cpe struct
	CPEs := []CPE{}

	// initialize CPEs and send bootstrap
	cpe1 := CPE{"1", "PIRELLI BROADBAND SOLUTIONS", "0013C8", "asd", "asd", "asd", "0 BOOTSTRAP"}
	cpe2 := CPE{"2", "Telsey", "0014", "asd", "asd", "asd", "1 BOOT"}

	CPEs = append(CPEs, cpe1)
	CPEs = append(CPEs, cpe2)

//	fmt.Println(CPEs)

	for _, c := range(CPEs) {
		go c.runConnection()
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

