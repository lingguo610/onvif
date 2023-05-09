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

/****************************************************************
请求媒体文件，媒体文件包含了一套媒体配置。
媒体文件被NVT的客户端媒体流配置属性所使用。
NVT在启动时提供至少有一个媒体文件。
*****************************************************************/
type ProfileRequest struct {
	XMLName string `xml:"trt:GetProfiles"`
}

type ProfileResponse struct {
	XMLName string    `xml:"Envelope"`
	Profile []Profile `xml:"Body>GetProfilesResponse>Profiles"`
}

type Profile struct {
	Name           string                      `xml:"Name"`
	Token          string                      `xml:"token,attr"` //文件令牌
	Fixed          bool                        `xml:"fixed,attr"`
	VideoConfig    VideoSourceConfiguration    `xml:"VideoSourceConfiguration"`    //视频源配置
	AudioConfig    AudioSourceConfiguration    `xml:"AudioSourceConfiguration"`    //音频源配置
	VideoEncoder   VideoEncoderConfiguration   `xml:"VideoEncoderConfiguration"`   //视频编码配置
	AudioEncoder   AudioEncoderConfiguration   `xml:"AudioEncoderConfiguration"`   //音频编码配置
	VideoAnalytics VideoAnalyticsConfiguration `xml:"VideoAnalyticsConfiguration"` //视频分析配置
	PTZ            PTZConfiguration            `xml:"PTZConfiguration"`            //云台配置
}

/******************************************************************
视频源配置
*******************************************************************/
type VideoSourceConfiguration struct {
	Token       string `xml:"token,attr"` //视频源令牌
	Name        string `xml:"Name"`
	UseCount    int    `xml:"UseCount"`
	SourceToken string `xml:"SourceToken"`
	Bounds      Bound  `xml:"Bounds"`
	Mode        string `xml:"Extension>Rotate>Mode"`
}

type Bound struct {
	X      int `xml:"x,attr"`
	Y      int `xml:"y,attr"`
	Width  int `xml:"width,attr"`
	Height int `xml:"height,attr"`
}

/******************************************************************
音频源配置
*******************************************************************/
type AudioSourceConfiguration struct {
	Token       string `xml:"token,attr"` //音频源令牌
	Name        string `xml:"Name"`
	UseCount    int    `xml:"UseCount"`
	SourceToken string `xml:"SourceToken"`
}

/******************************************************************
视频编码配置
*******************************************************************/
type VideoEncoderConfiguration struct {
	Token          string      `xml:"token,attr"`
	Name           string      `xml:"Name"`
	UseCount       int         `xml:"UseCount"`
	Encoding       string      `xml:"Encoding"`
	Resolution     Resolution  `xml:"Resolution"`
	Quality        string      `xml:"Quality"`
	RateControl    RateControl `xml:"RateControl"`
	H264           H264        `xml:"H264"`
	Multicast      Multicast   `xml:"Multicast"`
	SessionTimeout string      `xml:"SessionTimeout"`
}

type Resolution struct {
	Width  int `xml:"Width"`
	Height int `xml:"Height"`
}

type RateControl struct {
	FrameRateLimit   int `xml:"FrameRateLimit"`
	EncodingInterval int `xml:"EncodingInterval"`
	BitrateLimit     int `xml:"BitrateLimit"`
}

type H264 struct {
	GovLength   int    `xml:"GovLength"`
	H264Profile string `xml:"H264Profile"`
}

type Multicast struct {
	Address   string `xml:"Address"`
	Port      int    `xml:"Port"`
	TTL       int    `xml:"TTL"`
	AutoStart bool   `xml:"AutoStart"`
}

type Address struct {
	Type        string `xml:"Type"`
	IPv4Address string `xml:"IPv4Address"`
}

/******************************************************************
音频编码配置
*******************************************************************/
type AudioEncoderConfiguration struct {
	Token          string    `xml:"token,attr"`
	Name           string    `xml:"Name"`
	UseCount       int       `xml:"UseCount"`
	Encoding       string    `xml:"Encoding"`
	Bitrate        int       `xml:"Bitrate"`
	SampleRate     int       `xml:"SampleRate"`
	Multicast      Multicast `xml:"Multicast"`
	SessionTimeout string    `xml:"SessionTimeout"`
}

/******************************************************************
视频分析配置
*******************************************************************/
type VideoAnalyticsConfiguration struct {
	Token                        string                       `xml:"token,attr"`
	Name                         string                       `xml:"Name"`
	UseCount                     int                          `xml:"UseCount"`
	AnalyticsEngineConfiguration AnalyticsEngineConfiguration `xml:"AnalyticsEngineConfiguration"`
}

type AnalyticsEngineConfiguration struct {
	AnalyticsModule []AnalyticsModule `xml:"AnalyticsModule"`
}

//分析模块
type AnalyticsModule struct {
	Name        string      `xml:"Name,attr"`
	Type        string      `xml:"Type,attr"`
	SimpleItem  SimpleItem  `xml:"SimpleItem"`
	ElementItem ElementItem `xml:"ElementItem"`
}

type SimpleItem struct {
	Name  string `xml:"Name,attr"`
	Value int    `xml:"Value,attr"`
}

type ElementItem struct {
	Name       string     `xml:"Name,attr"`
	CellLayout CellLayout `xml:"CellLayout"`
}
type CellLayout struct {
	Columns   int `xml:"Columns,attr"`
	Rows      int `xml:"Rows,attr"`
	Translate XY  `xml:"Transformation>Translate"`
	Scale     XY  `xml:"Transformation>Scale"`
}

type XY struct {
	X string `xml:"x,attr"`
	Y string `xml:"y,attr"`
}

/******************************************************************
视频源配置
*******************************************************************/
type PTZConfiguration struct {
	Token                                 string `xml:"token,attr"`
	Name                                  string `xml:"Name"`
	UseCount                              int    `xml:"UseCount"`
	NodeToken                             string `xml:"NodeToken"`                             //节点Token
	DefaultContinuousPanTiltVelocitySpace string `xml:"DefaultContinuousPanTiltVelocitySpace"` //默认的连续的全方位速率空间
	DefaultContinuousZoomVelocitySpace    string `xml:"DefaultContinuousZoomVelocitySpace"`    //默认的连续的变焦速率空间
	DefaultPTZTimeout                     string `xml:"DefaultPTZTimeout"`                     //默认的移动
}

func (device *OnvifDevice) SetAuth(user, passwd, devIp string) {
	device.User = user
	device.Passwd = passwd
	device.DeviceIp = devIp
}

//请求NVT现有的媒体文件
func (device *OnvifDevice) GetProfiles() (*ProfileResponse, error) {

	if device.Capabilities == nil {
	log.Println("device.Capabilities == nil")
		if _, err := device.GetCapabilities(); err != nil {
		log.Println("device.GetCapabilities fail")
			return nil, errors.New("device.GetCapabilities fail")
		}
	}

	var profile ProfileRequest
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

	req.Header.Add("SOAPAction", "'http://www.onvif.org/ver10/media/wsdl/GetProfiles'")
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
	log.Println("1 device.GetCapabilities fail")
		return nil, err
	}
	
	log.Println("GetProfiles resp.StatusCode:", resp.StatusCode)
	
	
	if resp.StatusCode == 401 {
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
		A2 := fmt.Sprintf("POST:%s", "/onvif/media")
		io.WriteString(h, A2)
		HA2 := hex.EncodeToString(h.Sum(nil))

		// response
		cnonce := RandomKey()
		response := H(strings.Join([]string{HA1, nonceHeader, nc, cnonce, qopHeader, HA2}, ":"))

		// now make header
		AuthHeader := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s", qop=%s, nc=%s, cnonce="%s", opaque="%s", algorithm="%s"`,
			"ww", realmHeader, nonceHeader, "/onvif/media", response, qopHeader, nc, cnonce, opaqueHeader, algorithm)

		headers := http.Header{
			"User-Agent":      []string{userAgent},
			"Accept":          []string{"*/*"},
			"Accept-Encoding": []string{"identity"},
			"Connection":      []string{"Keep-Alive"},
			"Host":            []string{req.Host},
			"Authorization":   []string{AuthHeader},
		}
		req, err = http.NewRequest("POST", endpoint, httpbody)
	if err != nil {
		log.Println("http.NewRequest fail", err)
		return nil, err
	}

	req.Header.Add("SOAPAction", "'http://www.onvif.org/ver10/media/wsdl/GetProfiles'")
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
		req.Header = headers
		resp, err = client.Do(req)
		if err != nil {
			log.Println(" GetProfiles client.Do fail, err", err)
			return nil, err
		}
		defer resp.Body.Close()
	}
	
	log.Println("GetProfiles resp.StatusCode:", resp.StatusCode)
	
	if resp.StatusCode != 200 {
		log.Println("status code is not 200")
		return nil, errors.New("")
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ii := &ProfileResponse{}

	err = xml.Unmarshal(s, ii)
	if err != nil {
		log.Println("xml.Unmarshal fail", err)
		return nil, err
	}

	device.Profile = ii

	return ii, nil
}
