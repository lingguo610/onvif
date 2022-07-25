package device

import (
	"bytes"
	"errors"
	"log"
	"net/http"

	"github.com/lingguo610/onvif"
)

/****************************************************************
云台控制
*****************************************************************/

/******************************************************************
在某一个方向，某一个速度下，连续移动
*******************************************************************/
type ContinuousMoveRequest struct {
	XMLName      string   `xml:"tptz:ContinuousMove"`
	ProfileToken string   `xml:"ProfileToken"`
	Velocity     Velocity `xml:"Velocity"`
	Timeout      string   `xml:"Timeout"`
}

/******************************************************************
速度矢量，包括方向和长度
*******************************************************************/
type Velocity struct {
	PanTilt PanTilt `xml:"PanTilt"`
	Zoom    Zoom    `xml:"Zoom"`
}

type PanTilt struct {
	X     int    `xml:"x,attr"`
	Y     int    `xml:"y,attr"`
	Space string `xml:"space,attr"` //云台坐标，一个空间URI表示一个根本的坐标系统
}

type Zoom struct {
	X int `xml:"x,attr"`
}

type ContinuousMoveResponse struct {
	XMLName      string `xml:"Envelope"`
	Capabilities string `xml:"Body>ContinuousMoveResponse"`
}

/******************************************************************
连续的速率空间用于在一个指定方向上连续移动PTZ状态
VelocityGenericSpace 表示泛化的方位速率空间，它不涉及指定的物理范围。（左右上下移动）
ZoomSpaces_VelocityGenericSpace 表示泛化的变焦速率空间

*******************************************************************/
const (
	PanTiltSpaces_VelocityGenericSpace = "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace"
	ZoomSpaces_VelocityGenericSpace    = "http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace"
)

func (device *OnvifDevice) PTZContinuesMove(command onvif.CommandType) error {

	if device.Capabilities == nil {
		_, err := device.GetCapabilities()
		if err != nil {
			log.Println("get GetCapabilities fail")
			return errors.New("get GetCapabilities fail")
		}
	}

	ptzAddr := device.Capabilities.Capabilities.PTZ.XAddr
	if ptzAddr == "" {
		log.Println("the device do not support ptz")
		return errors.New("the device do not support ptz")
	}

	if device.Profile == nil {
		_, err := device.GetProfiles()
		if err != nil {
			log.Println("get profile fail")
			return errors.New("get profile fail")
		}
	}

	if len(device.Profile.Profile) <= 0 {
		return errors.New("the device has no profile")
	}
	token := device.Profile.Profile[0].Token

	var request ContinuousMoveRequest
	request.ProfileToken = token

	request.Velocity.PanTilt.Space = PanTiltSpaces_VelocityGenericSpace
	request.Timeout = "PT00H01M00S"
	request.Velocity.PanTilt.X = 0
	request.Velocity.PanTilt.Y = 0

	switch command {
	case onvif.LEFT:
		request.Velocity.PanTilt.X = 1
	case onvif.RIGHT:
		request.Velocity.PanTilt.X = -1
	case onvif.UP:
		request.Velocity.PanTilt.Y = 1
	case onvif.DOWN:
		request.Velocity.PanTilt.Y = -1
	default:
		request.Velocity.PanTilt.X = 0
	}

	element, err := buildElement(request)
	if err != nil {
		log.Println("buildElement profile fail")
		return errors.New("buildElement profile fail")
	}

	soap := NewEmptySOAP()
	soap.AddBodyContent(element)
	soap.AddWSSecurity(device.User, device.Passwd)
	httpbody := bytes.NewBufferString(soap.String())

	client := &http.Client{}
	req, err := http.NewRequest("POST", ptzAddr, httpbody)
	if err != nil {
		log.Println("http.NewRequest fail", err)
		return err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Add("SOAPAction", "'http://www.onvif.org/ver20/ptz/wsdl/ContinuousMove'")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("client.Do fail", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Println("status code is not 200")
		return errors.New("")
	}

	return nil
}
