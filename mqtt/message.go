package mqtt

// fixed header
// 4bit mqqt control packet ,
// 4bit flag specific to each mqtt control packet byte

import (
	"errors"
	_ "fmt"
)

const (
	RESERVED1 = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	RESERVED2
)

func parseVariableLen(buf []byte) (int, int, error) {
	multiplier := 1
	value := 0
	index := 0
	for _, b := range buf[:] {
		index += 1
		value += int(b&127) * multiplier
		if multiplier > 128*128*128 {
			return 0, 0, errors.New("malformed remain length")
		}
		if b&128 != 0 {
			break
		}
	}
	return value, index, nil
}

type Header struct {
	ControlFlag byte
	RemainLen   int32
	// variable header, with
	// PUBLISH QOS>0, PUBACK,PUBREC,PUBREL, PUBCOMP, SUBSCRIBE,SUBACK,UNSUBSCRIBE,UNSUBACK
	// Each time a Client sends a new packet of one of these types it MUST assign it a currently unused Packet Identifier
	// If a Client re-sends a particular Control Packet, then it MUST use the same Packet Identifier in subsequent re-sends of that packet
	PacketIdentifier int16
	Payload          []byte
}

func (header *Header) decodeHeader(buf []byte) (int, error) {
	header.ControlFlag = buf[0]
	// messageType := header.ControlFlag >> 4
	RemainLen, index, err := parseVariableLen(buf[1:])
	if err != nil {
		return 0, err
	}

	header.RemainLen = int32(RemainLen)
	return index + 1, nil
}

func (header *Header) MessageType() int {
	return int(header.ControlFlag >> 4)
}

func (header *Header) Flag() int {
	return int(header.ControlFlag & 0x07)
}

type Message interface {
	Encode() ([]byte, error)
	Decode([]byte)
	Verify() error
}

type ConnectMessage struct {
	Header
	Msb     byte
	Lsb     byte
	Proto   [4]byte
	Level   byte
	Connect byte
}

func (msg *ConnectMessage) Decode(buf []byte) error {
	_, err := msg.decodeHeader(buf)
	if err != nil {
		return err
	}

	// begin to parse remain length
	//fmt.Println(len)
	//

	return nil
}
