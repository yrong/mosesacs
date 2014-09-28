package cwmp

import (
	"encoding/xml"
	"strings"
)

type SoapEnvelope struct {
	XMLName xml.Name
	Header  SoapHeader
	Body    SoapBody
}

type SoapHeader struct{}
type SoapBody struct {
	CWMPMessage CWMPMessage `xml:",any"`
}

type CWMPMessage struct {
	XMLName xml.Name
}

type EventStruct struct {
	EventCode  string
	CommandKey string
}

type ParameterValueStruct struct {
	Name  string
	Value string
}

type ParameterInfoStruct struct {
	Name     string
	Writable string
}

type GetParameterValuesResponse struct {
	ParameterList []ParameterValueStruct `xml:"Body>GetParameterValuesResponse>ParameterList>ParameterValueStruct"`
}

type GetParameterNamesResponse struct {
	ParameterList []ParameterInfoStruct `xml:"Body>GetParameterNamesResponse>ParameterList>ParameterInfoStruct"`
}

type CWMPInform struct {
	DeviceId      DeviceID               `xml:"Body>Inform>DeviceId"`
	Events        []EventStruct          `xml:"Body>Inform>Event>EventStruct"`
	ParameterList []ParameterValueStruct `xml:"Body>Inform>ParameterList>ParameterValueStruct"`
}

func (s *SoapEnvelope) KindOf() string {
	return s.Body.CWMPMessage.XMLName.Local
}

func (i *CWMPInform) GetEvents() string {
	res := ""
	for idx := range i.Events {
		res += i.Events[idx].EventCode
	}

	return res
}

func (i *CWMPInform) GetConnectionRequest() string {
	for idx := range i.ParameterList {
		// valid condition for both tr98 and tr181
		if strings.HasSuffix(i.ParameterList[idx].Name,"Device.ManagementServer.ConnectionRequestURL") {
			return i.ParameterList[idx].Value
		}
	}

	return ""
}

func (i *CWMPInform) GetSoftwareVersion() string {
	for idx := range i.ParameterList {
		if strings.HasSuffix(i.ParameterList[idx].Name,"Device.DeviceInfo.SoftwareVersion") {
			return i.ParameterList[idx].Value
		}
	}

	return ""
}

func (i *CWMPInform) GetHardwareVersion() string {
	for idx := range i.ParameterList {
		if strings.HasSuffix(i.ParameterList[idx].Name,"Device.DeviceInfo.HardwareVersion") {
			return i.ParameterList[idx].Value
		}
	}

	return ""
}

func (i *CWMPInform) GetDataModelType() string {
	if 	strings.HasPrefix(i.ParameterList[0].Name,"InternetGatewayDevice") {
		return "TR098"
	} else if strings.HasPrefix(i.ParameterList[0].Name,"Device") {
		return "TR181"
	}

	return ""
}

type DeviceID struct {
	Manufacturer string
	OUI          string
	SerialNumber string
}

func InformResponse() string {
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

func GetParameterValues(leaf string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:schemaLocation="urn:dslforum-org:cwmp-1-0 ..\schemas\wt121.xsd" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soap:Header/>
  <soap:Body soap:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
    <cwmp:GetParameterValues>
      <ParameterNames>
      	<string>` + leaf + `</string>
      </ParameterNames>
    </cwmp:GetParameterValues>
  </soap:Body>
</soap:Envelope>`
}

func GetParameterNames(leaf string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:schemaLocation="urn:dslforum-org:cwmp-1-0 ..\schemas\wt121.xsd" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soap:Header/>
  <soap:Body soap:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
    <cwmp:GetParameterNames>
      <ParameterPath>` + leaf + `</ParameterPath>
      <NextLevel>1</NextLevel>
    </cwmp:GetParameterNames>
  </soap:Body>
</soap:Envelope>`
}

func FactoryReset() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soapenc="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:cwmp="urn:dslforum-org:cwmp-1-0" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:schemaLocation="urn:dslforum-org:cwmp-1-0 ..\schemas\wt121.xsd" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soap:Header/>
  <soap:Body soap:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
    <cwmp:FactoryReset/>
  </soap:Body>
</soap:Envelope>`
}
