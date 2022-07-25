package internal

import (
	"encoding/xml"
	"log"

	"github.com/beevik/etree"
)

type OnvifDevice struct {
	User     string
	Passwd   string
	DeviceIp string

	Profile      *ProfileResponse
	StreamUri    *StreamUriResponse
	Capabilities *CapbilityResponse
}

func buildElement(method interface{}) (*etree.Element, error) {
	output, err := xml.MarshalIndent(method, "  ", "    ")
	if err != nil {
		log.Printf("error: %v\n", err.Error())
		return nil, err
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(string(output)); err != nil {
		return nil, err
	}
	element := doc.Root()
	return element, nil
}
