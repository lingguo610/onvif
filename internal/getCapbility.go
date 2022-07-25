package internal

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type CapbilityRequest struct {
	XMLName  string `xml:"tds:GetCapabilities"`
	Category string `xml:"Category"`
}

type CapbilityResponse struct {
	XMLName      string       `xml:"Envelope"`
	Capabilities Capabilities `xml:"Body>GetCapabilitiesResponse>Capabilities"`
}

type Capabilities struct {
	Analytics Analytics `xml:"Analytics"`
	Device    Device    `xml:"Device"`
	Events    Events    `xml:"Events"`
	Imaging   Imaging   `xml:"Imaging"`
	Media     Media     `xml:"Media"`
	PTZ       PTZ       `xml:"PTZ"`
}

type Analytics struct {
	XAddr                  string `xml:"XAddr"`
	RuleSupport            bool   `xml:"RuleSupport"`
	AnalyticsModuleSupport bool   `xml:"AnalyticsModuleSupport"`
}

type Device struct {
	XAddr             string `xml:"XAddr"`
	IPFilter          bool   `xml:"Network>IPFilter"`
	ZeroConfiguration bool   `xml:"Network>ZeroConfiguration"`
	IPVersion6        bool   `xml:"Network>IPVersion6"`
	DynDNS            bool   `xml:"Network>DynDNS"`

	DiscoveryResolve  bool                      `xml:"System>DiscoveryResolve"`
	DiscoveryBye      bool                      `xml:"System>DiscoveryBye"`
	RemoteDiscovery   bool                      `xml:"System>RemoteDiscovery"`
	SystemBackup      bool                      `xml:"System>SystemBackup"`
	SystemLogging     bool                      `xml:"System>SystemLogging"`
	FirmwareUpgrade   bool                      `xml:"System>FirmwareUpgrade"`
	SupportedVersions []SystemSupportedVersions `xml:"System>SupportedVersions"`

	InputConnectors int `xml:"IO>InputConnectors"`
	RelayOutputs    int `xml:"IO>RelayOutputs"`

	TLS11                bool `xml:"Security>TLS1.1"`
	TLS12                bool `xml:"Security>TLS1.2"`
	OnboardKeyGeneration bool `xml:"Security>OnboardKeyGeneration"`
	AccessPolicyConfig   bool `xml:"Security>AccessPolicyConfig"`
	X509Token            bool `xml:"Security>X.509Token"`
	SAMLToken            bool `xml:"Security>SAMLToken"`
	KerberosToken        bool `xml:"Security>KerberosToken"`
	RELToken             bool `xml:"Security>RELToken"`
}

type SystemSupportedVersions struct {
	Major int `xml:"Major"`
	Minor int `xml:"Minor"`
}

type Events struct {
	XAddr                                         string `xml:"XAddr"`
	WSSubscriptionPolicySupport                   bool   `xml:"WSSubscriptionPolicySupport"`
	WSPullPointSupport                            bool   `xml:"WSPullPointSupport"`
	WSPausableSubscriptionManagerInterfaceSupport bool   `xml:"WSPausableSubscriptionManagerInterfaceSupport"`
}

type Imaging struct {
	XAddr string `xml:"XAddr"`
}

type Media struct {
	XAddr        string `xml:"XAddr"`
	RTPMulticast bool   `xml:"StreamingCapabilities>RTPMulticast"`
	RTP_TCP      bool   `xml:"StreamingCapabilities>RTP_TCP"`
	RTP_RTSP_TCP bool   `xml:"StreamingCapabilities>RTP_RTSP_TCP"`
}

type PTZ struct {
	XAddr string `xml:"XAddr"`
}

func (device *OnvifDevice) GetCapabilities() (*CapbilityResponse, error) {

	var request CapbilityRequest
	request.Category = "All"
	element, err := buildElement(request)
	if err != nil {
		log.Println("buildElement profile fail")
		return nil, errors.New("buildElement profile fail")
	}

	soap := NewEmptySOAP()
	soap.AddBodyContent(element)
	httpbody := bytes.NewBufferString(soap.String())

	endpoint := "http://" + device.DeviceIp + "/onvif/device_service"
	client := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, httpbody)
	if err != nil {
		log.Println("http.NewRequest fail", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Println("status code is not 200")
		return nil, errors.New("")
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ii := &CapbilityResponse{}

	err = xml.Unmarshal(s, ii)
	if err != nil {
		log.Println("xml.Unmarshal fail", err)
		return nil, err
	}

	device.Capabilities = ii

	return ii, nil
}
