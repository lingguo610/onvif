package device

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type StreamUriRequest struct {
	XMLName      string `xml:"trt:GetStreamUri"`
	Stream       string `xml:"StreamSetup>Stream"`
	Transport    string `xml:"StreamSetup>Transport"`
	ProfileToken string `xml:"ProfileToken"`
}

type StreamUriResponse struct {
	XMLName  string   `xml:"Envelope"`
	MediaUri MediaUri `xml:"Body>GetStreamUriResponse>MediaUri"`
}

type MediaUri struct {
	Uri                   string `xml:"Uri"`
	InvalidAfterConnected bool   `xml:"InvalidAfterConnect"`
	InvalidAfterReboot    bool   `xml:"InvalidAfterReboot"`
	TimeOut               string `xml:"Timeout"`
}

func (device *OnvifDevice) getStreamUri() (*StreamUriResponse, error) {

	if device.Profile == nil {
		_ , err := device.GetProfiles()
		if err != nil {
			return nil, errors.New("get profile fail")
		}
	}

	if len(device.Profile.Profile) <= 0 {
		return nil, errors.New("the device has no profile")
	}
	token := device.Profile.Profile[0].Token

	var profile StreamUriRequest
	profile.ProfileToken = token
	profile.Stream = "RTP-Unicast"
	profile.Transport = "UDP"
	element, err := buildElement(profile)
	if err != nil {
		log.Println("buildElement profile fail")
		return nil, errors.New("buildElement profile fail")
	}

	soap := NewEmptySOAP()
	soap.AddBodyContent(element)
	soap.AddWSSecurity(device.User, device.Passwd)
	httpbody := bytes.NewBufferString(soap.String())

	//endpoint := "http://" + device.DeviceIp + "/onvif/media"
	endpoint := device.Capabilities.Capabilities.Media.XAddr
	client := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, httpbody)
	if err != nil {
		log.Println("http.NewRequest fail", err)
		return nil, err
	}

	req.Header.Add("SOAPAction", "'http://www.onvif.org/ver10/media/wsdl/GetStreamUri'")
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ii := &StreamUriResponse{}

	err = xml.Unmarshal(s, ii)
	if err != nil {
		log.Println("xml.Unmarshal fail", err)
		return nil, err
	}

	device.StreamUri = ii

	return ii, nil
}

func (device *OnvifDevice) GetMediaUri() (string, error) {
	if device.StreamUri == nil {
		device.getStreamUri()
	}

	if device.StreamUri == nil {
		return "", errors.New("get StreamUri fail")
	}

	return device.StreamUri.MediaUri.Uri, nil

}
