package cwmpclient

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"bytes"
	"github.com/lucacervasio/mosesacs/cwmp"
	"net/url"
	"encoding/xml"
	"io/ioutil"
	"time"
	"math/rand"
	"strconv"
	"strings"
)

var acs_url string
var serial string

func connectionRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got connection request, send Inform to %s", acs_url)
	sendInform(strings.Split(r.RequestURI, "-")[1])
}

func random(min, max int) int {
	rand.Seed(int64(time.Now().Nanosecond()))
	return rand.Intn(max - min) + min
}

func RunClient(n int) {
	acs_url = "http://localhost:9292/acs"

	for i := 0; i < n; i++ {
		serial = strconv.Itoa(random(1000,5000))
		connection_request_url := "/ConnectionRequest-"+serial
		log.Println(connection_request_url)
		http.HandleFunc(connection_request_url, connectionRequestHandler)
	}

	go func() {
		log.Println("Start http server waiting connection request")
		http.ListenAndServe(":7547", nil)

	}()
}

func sendInform(s string){
	log.Printf("send Inform to %s", acs_url)
	var msgToSend []byte
	msgToSend = []byte(Inform(s))

	tr := &http.Transport{}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Transport: tr, Jar:jar}
	envelope := cwmp.SoapEnvelope{}
	u, _ := url.Parse(acs_url)

	resp, _ := client.Post(acs_url, "text/xml", bytes.NewBuffer(msgToSend))
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
				log.Println(body)
				var e cwmp.GetParameterValues_
				xml.Unmarshal([]byte(body), &e)
				log.Println(e)
				msgToSend = []byte(GetParameterValuesResponse(s))
			} else {
				log.Println("send empty post")
				msgToSend = []byte("")
			}

			client.Jar.SetCookies(u, resp.Cookies())
			resp, _ = client.Post(acs_url, "text/xml", bytes.NewBuffer(msgToSend))
			log.Println(resp.Header)
		}
	}

}

func Inform(s string) string {
	return `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:cwmp="urn:dslforum-org:cwmp-1-0"><soap:Header><cwmp:ID soap:mustUnderstand="1">5058</cwmp:ID></soap:Header>
	<soap:Body><cwmp:Inform><DeviceId><Manufacturer>ADB Broadband</Manufacturer>
<OUI>0013C8</OUI>
<ProductClass>VV5522</ProductClass>
<SerialNumber>PI234550701S199991-`+ s +`</SerialNumber>
</DeviceId>
<Event soap-enc:arrayType="cwmp:EventStruct[1]">
<EventStruct><EventCode>6 CONNECTION REQUEST</EventCode>
<CommandKey></CommandKey>
</EventStruct>
</Event>
<MaxEnvelopes>1</MaxEnvelopes>
<CurrentTime>` + time.Now().Format(time.RFC3339) + `</CurrentTime>
<RetryCount>0</RetryCount>
<ParameterList soap-enc:arrayType="cwmp:ParameterValueStruct[8]">
<ParameterValueStruct><Name>InternetGatewayDevice.ManagementServer.ConnectionRequestURL</Name>
<Value xsi:type="xsd:string">http://localhost:7547/ConnectionRequest-`+s+`</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.ManagementServer.ParameterKey</Name>
<Value xsi:type="xsd:string"></Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceSummary</Name>
<Value xsi:type="xsd:string">InternetGatewayDevice:1.2[](Baseline:1,EthernetLAN:1,WiFiLAN:1,ADSLWAN:1,EthernetWAN:1,QoS:1,QoSDynamicFlow:1,Bridging:1,Time:1,IPPing:1,TraceRoute:1,DeviceAssociation:1,UDPConnReq:1),VoiceService:1.0[1](TAEndpoint:1,SIPEndpoint:1)</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.HardwareVersion</Name>
<Value xsi:type="xsd:string">`+s+`</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.ProvisioningCode</Name>
<Value xsi:type="xsd:string">ABCD</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.SoftwareVersion</Name>
<Value xsi:type="xsd:string">E_8.0.0.0002</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.SpecVersion</Name>
<Value xsi:type="xsd:string">1.0</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.ExternalIPAddress</Name>
<Value xsi:type="xsd:string">12.0.0.10</Value>
</ParameterValueStruct>
</ParameterList>
</cwmp:Inform>
</soap:Body></soap:Envelope>`
}

func GetParameterValuesResponse(s string) string {
	return `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:soap-enc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:cwmp="urn:dslforum-org:cwmp-1-0">
	<soap:Header><cwmp:ID soap:mustUnderstand="1">3</cwmp:ID></soap:Header>
	<soap:Body><cwmp:GetParameterValuesResponse><ParameterList soap-enc:arrayType="cwmp:ParameterValueStruct[20]">
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.AdditionalHardwareVersion</Name>
<Value xsi:type="xsd:string"></Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.AdditionalSoftwareVersion</Name>
<Value xsi:type="xsd:string">E_8.0.0.0002</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.Description</Name>
<Value xsi:type="xsd:string"></Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.HardwareVersion</Name>
<Value xsi:type="xsd:string">VV5522</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.Manufacturer</Name>
<Value xsi:type="xsd:string">ADB Broadband</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.ManufacturerOUI</Name>
<Value xsi:type="xsd:string">0013C8</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.ModelName</Name>
<Value xsi:type="xsd:string">`+s+`</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.ProvisioningCode</Name>
<Value xsi:type="xsd:string">ABCD</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.SerialNumber</Name>
<Value xsi:type="xsd:string">PI234550701S199991-VV5522</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.SoftwareVersion</Name>
<Value xsi:type="xsd:string">E_8.0.0.0002</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.SpecVersion</Name>
<Value xsi:type="xsd:string">1.0</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.VendorConfigFileNumberOfEntries</Name>
<Value xsi:type="xsd:unsignedInt">1</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.ProductClass</Name>
<Value xsi:type="xsd:string">VV5522</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.FirstUseDate</Name>
<Value xsi:type="xsd:dateTime">2013-10-15T15:40:33Z</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.VendorConfigFile.1.Name</Name>
<Value xsi:type="xsd:string">multi_user</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.VendorConfigFile.1.Version</Name>
<Value xsi:type="xsd:string">E_8.0.0.0002</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.VendorConfigFile.1.Date</Name>
<Value xsi:type="xsd:dateTime">Tue Oct 15 15:48:15 UTC 2013</Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.VendorConfigFile.1.Description</Name>
<Value xsi:type="xsd:string">multi_user</Value>
<Value xsi:type="xsd:string"></Value>
</ParameterValueStruct>
<ParameterValueStruct><Name>InternetGatewayDevice.DeviceInfo.UpTime</Name>
<Value xsi:type="xsd:unsignedInt">5062</Value>
</ParameterValueStruct>
</ParameterList>
</cwmp:GetParameterValuesResponse>
</soap:Body></soap:Envelope>`
}