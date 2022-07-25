package device

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"time"

	"github.com/elgs/gostrgen"
)

type wsAuth struct {
	XMLName  xml.Name `xml:"wsse:UsernameToken"`
	Username string   `xml:"wsse:Username"`
	Password password `xml:"wsse:Password"`
	Nonce    nonce    `xml:"wsse:Nonce"`
	Created  string   `xml:"wsse:Created"`
}

type password struct {
	Type     string `xml:"Type,attr"`
	Password string `xml:",chardata"`
}

type nonce struct {
	Type  string `xml:"EncodingType,attr"`
	Nonce string `xml:",chardata"`
}

type Security struct {
	XMLName xml.Name `xml:"wsse:Security"`
	Auth    wsAuth
}

const (
	passwordType = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest"
	encodingType = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"
)

func NewSecurity(username, passwd string) Security {

	charsToGenerate := 32
	charSet := gostrgen.Lower | gostrgen.Digit

	nonceSeq, _ := gostrgen.RandGen(charsToGenerate, charSet, "", "")
	created := time.Now().UTC().Format(time.RFC3339Nano)
	auth := Security{
		Auth: wsAuth{
			Username: username,
			Password: password{
				Type:     passwordType,
				Password: generateToken(username, nonceSeq, created, passwd),
			},
			Nonce: nonce{
				Type:  encodingType,
				Nonce: nonceSeq,
			},
			Created: created,
		},
	}

	return auth
}

func generateToken(Username string, Nonce string, Created string, Password string) string {
	sDec, _ := base64.StdEncoding.DecodeString(Nonce)

	hasher := sha1.New()
	hasher.Write([]byte(string(sDec) + Created + Password))

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
