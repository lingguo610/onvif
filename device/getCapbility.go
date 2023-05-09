package device

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"crypto/md5"
	"fmt"
	"io"
	"encoding/hex"
	"strings"
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
	
	if resp.StatusCode == 401 {
		log.Println("resp.StatusCode == 401")
		var authorization map[string]string = DigestAuthParams(resp)
		realmHeader := authorization["realm"]
		qopHeader := authorization["qop"]
		nonceHeader := authorization["nonce"]
		opaqueHeader := authorization["opaque"]
		algorithm := authorization["algorithm"]
		realm := realmHeader
		// A1
		h := md5.New()
		A1 := fmt.Sprintf("%s:%s:%s", "ww", realm, "123456ab")
		io.WriteString(h, A1)
		HA1 := hex.EncodeToString(h.Sum(nil))

		// A2
		h = md5.New()
		A2 := fmt.Sprintf("GET:%s", "/auth")
		io.WriteString(h, A2)
		HA2 := hex.EncodeToString(h.Sum(nil))

		// response
		cnonce := RandomKey()
		response := H(strings.Join([]string{HA1, nonceHeader, nc, cnonce, qopHeader, HA2}, ":"))

		// now make header
		AuthHeader := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s", qop=%s, nc=%s, cnonce="%s", opaque="%s", algorithm="%s"`,
			"ww", realmHeader, nonceHeader, "/auth", response, qopHeader, nc, cnonce, opaqueHeader, algorithm)

		headers := http.Header{
			"User-Agent":      []string{userAgent},
			"Accept":          []string{"*/*"},
			"Accept-Encoding": []string{"identity"},
			"Connection":      []string{"Keep-Alive"},
			"Host":            []string{req.Host},
			"Authorization":   []string{AuthHeader},
		}
		//req, err = http.NewRequest("GET", uri, nil)
		req.Header = headers
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}
	
	log.Println("GetCapabilities resp.StatusCode:", resp.StatusCode)
	
	
	if resp.StatusCode != 200 {
		log.Println("GetCapabilities status code is not 200")
		return nil, errors.New("")
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	log.Println("GetCapabilities ioutil.ReadAll fail")
		return nil, err
	}

	ii := &CapbilityResponse{}

	err = xml.Unmarshal(s, ii)
	if err != nil {
		log.Println("GetCapabilities xml.Unmarshal fail", err)
		return nil, err
	}
	
	log.Println("GetCapabilities sucess")

	device.Capabilities = ii

	return ii, nil
}
