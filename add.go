package ldap

import (
	"fmt"
	"github.com/go-asn1-ber/asn1-ber"
	"log"
)

type AddRequest struct {
	Entry    *Entry
	Controls []Control
}

/*
func (req *AddRequest) RecordType() uint8 {
	return AddRecord
}
*/

func (l *Connection) Add(req *AddRequest) error {
	messageID, ok := l.nextMessageID()
	if !ok {
		return newError(ErrorClosing, "messageID channel is closed.")
	}

	if l.Debug {
		log.Println("sdfsadfsdf")
	}


	encodedAdd, err := encodeAddRequest(req)
	if err != nil {
		return err
	}

	if l.Debug {
		log.Println("encoded add request:", encodedAdd)
	}

	packet, err := requestBuildPacket(messageID, encodedAdd, req.Controls)
	if err != nil {
		return err
	}

	return l.sendReqRespPacket(messageID, packet)
}

/*
   AddRequest ::= [APPLICATION 8] SEQUENCE {
        entry           LDAPDN,
        attributes      AttributeList }

   AttributeList ::= SEQUENCE OF attribute Attribute

   Attribute ::= SEQUENCE {
        type       AttributeDescription,
        vals       SET OF value AttributeValue } // vals is not empty
*/
func encodeAddRequest(addReq *AddRequest) (*ber.Packet, error) {
	addPacket := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ber.Tag(ApplicationAddRequest), nil, ApplicationAddRequest.String())
	addPacket.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, addReq.Entry.DN, "LDAP DN"))

	attributeList := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "AttributeList")

	for _, attr := range addReq.Entry.Attributes {
		attribute := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Attribute")
		attribute.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, attr.Name, "Attribute Desc"))
		if len(attr.Values) == 0 {
			return nil, newError(ErrorEncoding, "attribute "+attr.Name+" had no values.")
		}
		valuesSet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "Attribute Value Set")
		for _, val := range attr.Values {
			valuesSet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, val, "AttributeValue"))
		}
		attribute.AppendChild(valuesSet)
		attributeList.AppendChild(attribute)
	}
	addPacket.AppendChild(attributeList)
	return addPacket, nil
}

func (req *AddRequest) Bytes() []byte {
	encoded, _ := encodeAddRequest(req)
	return encoded.Bytes()
}

func NewAddRequest(dn string) (req *AddRequest) {
	req = &AddRequest{Entry: NewEntry(dn), Controls: make([]Control, 0)}
	return
}

func (req *AddRequest) AddAttribute(attr *EntryAttribute) {
	req.Entry.AddAttributeValues(attr.Name, attr.Values)
}

func (req *AddRequest) AddAttributes(attrs []*EntryAttribute) {
	for _, attr := range attrs {
		req.Entry.AddAttributeValues(attr.Name, attr.Values)
	}
}

// DumpAddRequest - Basic LDIF "like" dump for testing, no formating, etc
func (addReq *AddRequest) String() (dump string) {
	dump = fmt.Sprintf("dn: %s\n", addReq.Entry.DN)
	for _, attr := range addReq.Entry.Attributes {
		for _, val := range attr.Values {
			dump += fmt.Sprintf("%s: %s\n", attr.Name, val)
		}
	}
	dump += fmt.Sprintf("\n")
	return
}

func (req *AddRequest) AddControl(control Control) {
	if req.Controls == nil {
		req.Controls = make([]Control, 0)
	}
	req.Controls = append(req.Controls, control)
}
