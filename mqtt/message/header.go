package message

// fixed header
// 4bit mqqt control packet ,
// 4bit flag specific to each mqtt control packet byte

import (
	"errors"
)

// in mqtt protocol, the message is seperated into three parts:
// -- Fixed header, present in all MQTT Control Packets
// -- Variable header, present in some MQTT Control Packets
// -- Payload, present in some MQTT Control Packets

// FixedHeader FixHeader
// is separated into
// | ------------------------|----------|
// | MQTT Control Packet type|Flags specific to each MQTT Control Packet type|
// | Remaining Length                     |
type FixedHeader struct {
	ControlFlag byte
	remainLen   uint32
}

// SetRemainLen set remain message length
func (header *FixedHeader) SetRemainLen(remainLen uint32) {
	header.remainLen = remainLen
}

func (header *FixedHeader) headerLen() int {
	// for first byte
	return 1 + vintLen(header.remainLen)

}
func (header *FixedHeader) encodeHeader(dest []byte) (int, error) {
	if len(dest) < header.headerLen() {
		return 0, errors.New("buff not have enough space=")
	}
	index := 0
	dest[index] = header.ControlFlag
	index++

	l := writeVint(dest[index:], header.remainLen)
	index += l
	return index, nil
}

//decodeHeader decode message header=
func (header *FixedHeader) decodeHeader(buf []byte) (int, error) {

	if len(buf) < 1 {
		return 0, errors.New("wrong message format")
	}
	header.ControlFlag = buf[0]
	// messageType := header.ControlFlag >> 4
	remainLen, l, err := readVint(buf[1:])
	if err != nil {
		return 0, err
	}

	header.remainLen = uint32(remainLen)
	return l + 1, nil
}

// MessageType return the message type, like connect connectack
func (header *FixedHeader) MessageType() byte {
	return header.ControlFlag >> 4
}

// SetMessageType set message type
func (header *FixedHeader) SetMessageType(t byte) error {
	if t < CONNECT || t > DISCONNECT {
		return errors.New("Invalid message type")
	}
	header.ControlFlag = (header.ControlFlag & 0xf) | (t&0xf)<<4
	return nil
}

// Flag return message flag
func (header *FixedHeader) Flag() int {
	return int(header.ControlFlag & 0xf)
}
