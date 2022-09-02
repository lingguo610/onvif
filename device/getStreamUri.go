package device

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type StreamUriRequest struct {
	XMLName      string `xml:"trt:GetStreamUri"`
	Stream       string `xml:"trt:StreamSetup>tt:Stream"`
	Transport    string `xml:"trt:StreamSetup>tt:Transport>tt:Protocol"`
	ProfileToken string `xml:"trt:ProfileToken"`
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
		_, err := device.GetProfiles()
		if err != nil {
			return nil, errors.New("get profile fail")
		}
	}

	if device.Profile == nil {
		return nil, errors.New("device.Profile == nil")
	}

	if len(device.Profile.Profile) <= 0 {
		return nil, errors.New("len(device.Profile.Profile) <= 0")
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

	fmt.Println("s: ", string(s))

	ii := &StreamUriResponse{}

	err = xml.Unmarshal(s, ii)
	if err != nil {
		log.Println("xml.Unmarshal fail", err)
		return nil, err
	}
	fmt.Println("ddd: ", ii)

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
