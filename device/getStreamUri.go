package device

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
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

const (
	nc        = "00000001"
	userAgent = "AtScale"
)

func (device *OnvifDevice) getStreamUri() (*StreamUriResponse, error) {

log.Println("enter getStreamUri")
log.Println("enter getStreamUri2")
	if device.Profile == nil {
	
	log.Println("device.Profile == nil")
		_, err := device.GetProfiles()
		if err != nil {
		log.Println("device.GetProfiles fail")
			return nil, errors.New("get profile fail")
		}
	}
	
	log.Println("11111")

	if device.Profile == nil {
	log.Println("device.Profile == nil")
		return nil, errors.New("device.Profile == nil")
	}
	
	log.Println("22222")

	if len(device.Profile.Profile) <= 0 {
	log.Println("len(device.Profile.Profile) <= 0")
		return nil, errors.New("len(device.Profile.Profile) <= 0")
	}
	
	log.Println("33333")
	
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
	log.Println("start NewRequest")

	soap := NewEmptySOAP()
	soap.AddBodyContent(element)
	soap.AddWSSecurity(device.User, device.Passwd)
	httpbody := bytes.NewBufferString(soap.String())

	endpoint := device.Capabilities.Capabilities.Media.XAddr
	client := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, httpbody)
	if err != nil {
		log.Println("http.NewRequest fail", err)
		return nil, err
	}

	req.Header.Add("SOAPAction", "'http://www.onvif.org/ver10/media/wsdl/GetStreamUri'")
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	log.Println("start client.Do")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	log.Println("resp.StatusCode:", resp.StatusCode)

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

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("getStreamUri response: ", string(s))

	ii := &StreamUriResponse{}

	err = xml.Unmarshal(s, ii)
	if err != nil {
		log.Println("xml.Unmarshal fail", err)
		return nil, err
	}
	fmt.Println("streamUri: ", ii)

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

func DigestAuthParams(r *http.Response) map[string]string {
	s := strings.SplitN(r.Header.Get("Www-Authenticate"), " ", 2)
	if len(s) != 2 || s[0] != "Digest" {
		return nil
	}

	result := map[string]string{}
	for _, kv := range strings.Split(s[1], ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.Trim(parts[0], "\" ")] = strings.Trim(parts[1], "\" ")
	}
	return result
}
func RandomKey() string {
	k := make([]byte, 8)
	for bytes := 0; bytes < len(k); {
		n, err := rand.Read(k[bytes:])
		if err != nil {
			panic("rand.Read() failed")
		}
		bytes += n
	}
	return base64.StdEncoding.EncodeToString(k)
}

/*
 H function for MD5 algorithm (returns a lower-case hex MD5 digest)
*/
func H(data string) string {
	digest := md5.New()
	digest.Write([]byte(data))
	return hex.EncodeToString(digest.Sum(nil))
}

func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		if rwTimeout > 0 {
			conn.SetDeadline(time.Now().Add(rwTimeout))
		}
		return conn, nil
	}
}
