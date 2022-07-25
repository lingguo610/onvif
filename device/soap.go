package device

import (
	"encoding/xml"
	"errors"
	"log"

	"github.com/beevik/etree"
)

type SoapMessage string

func NewEmptySOAP() SoapMessage {
	doc := buildSoapRoot()
	res, _ := doc.WriteToString()

	return SoapMessage(res)
}

func buildSoapRoot() *etree.Document {

	doc := etree.NewDocument()

	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	env := doc.CreateElement("soap-env:Envelope")
	env.CreateElement("soap-env:Header")
	env.CreateElement("soap-env:Body")

	env.CreateAttr("xmlns:soap-env", "http://www.w3.org/2003/05/soap-envelope")
	env.CreateAttr("xmlns:soap-enc", "http://www.w3.org/2003/05/soap-encoding")
	env.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	env.CreateAttr("xmlns:xsd", "http://www.w3.org/2001/XMLSchema")
	env.CreateAttr("xmlns:wsa", "http://schemas.xmlsoap.org/ws/2004/08/addressing")
	env.CreateAttr("xmlns:wsdd", "http://schemas.xmlsoap.org/ws/2005/04/discovery")
	env.CreateAttr("xmlns:c14n", "http://www.w3.org/2001/10/xml-exc-c14n#")
	env.CreateAttr("xmlns:ds", "http://www.w3.org/2000/09/xmldsig#")
	env.CreateAttr("xmlns:saml1", "urn:oasis:names:tc:SAML:1.0:assertion")
	env.CreateAttr("xmlns:saml2", "urn:oasis:names:tc:SAML:2.0:assertion")
	env.CreateAttr("xmlns:wsu", "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd")
	env.CreateAttr("xmlns:xenc", "http://www.w3.org/2001/04/xmlenc#")
	env.CreateAttr("xmlns:wsc", "http://docs.oasis-open.org/ws-sx/ws-secureconversation/200512")
	env.CreateAttr("xmlns:wsse", "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd")
	env.CreateAttr("xmlns:xmime", "http://tempuri.org/xmime.xsd")
	env.CreateAttr("xmlns:xop", "http://www.w3.org/2004/08/xop/include")
	env.CreateAttr("xmlns:wsa5", "http://www.w3.org/2005/08/addressing")
	env.CreateAttr("xmlns:wstop", "http://docs.oasis-open.org/wsn/t-1")
	env.CreateAttr("xmlns:tt", "http://www.onvif.org/ver10/schema")
	env.CreateAttr("xmlns:wsrfbf", "http://docs.oasis-open.org/wsrf/bf-2")
	env.CreateAttr("xmlns:wsrfr", "http://docs.oasis-open.org/wsrf/r-2")
	env.CreateAttr("xmlns:tdn", "http://www.onvif.org/ver10/network/wsdl")
	env.CreateAttr("xmlns:tds", "http://www.onvif.org/ver10/device/wsdl")
	env.CreateAttr("xmlns:tev", "http://www.onvif.org/ver10/events/wsdl")
	env.CreateAttr("xmlns:wsnt", "http://docs.oasis-open.org/wsn/b-2")
	env.CreateAttr("xmlns:tmd", "http://www.onvif.org/ver10/deviceIO/wsdl")
	env.CreateAttr("xmlns:tptz", "http://www.onvif.org/ver20/ptz/wsdl")
	env.CreateAttr("xmlns:tr2", "http://www.onvif.org/ver20/media/wsdl")
	env.CreateAttr("xmlns:trt", "http://www.onvif.org/ver10/media/wsdl")

	return doc
}

func (msg SoapMessage) String() string {
	return string(msg)
}

func (msg *SoapMessage) AddBodyContent(element *etree.Element) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(msg.String()); err != nil {
		log.Println(err.Error())
		return
	}

	bodyTag := doc.Root().SelectElement("Body")
	if bodyTag == nil {
		log.Println("body element is nil")
		return
	}

	bodyTag.AddChild(element)

	res, _ := doc.WriteToString()

	*msg = SoapMessage(res)
}

func (msg *SoapMessage) AddWSSecurity(user, passwd string) error {

	auth := NewSecurity(user, passwd)
	soapReq, err := xml.MarshalIndent(auth, "", "  ")
	if err != nil {
		return err
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(string(soapReq)); err != nil {
		return err
	}
	element := doc.Root()

	doc = etree.NewDocument()
	if err := doc.ReadFromString(msg.String()); err != nil {
		return err
	}

	bodyTag := doc.Root().SelectElement("Header")
	if bodyTag == nil {
		log.Println("body element is nil")
		return errors.New("header element is nil")
	}
	bodyTag.AddChild(element)

	res, _ := doc.WriteToString()

	*msg = SoapMessage(res)
	return nil
}
