package message

import "errors"

// ConnAck has 2 parts fixed header and variable header
// variable header is separated into two parts:
// Connect Acknowledge Flags 1
// Connect Return code 1
//

const (
	//ConnAccepted successful build connection
	ConnAccepted = iota

	//UnAcceptableVersion 0x01 Connection Refused, unacceptable protocol version
	UnAcceptableVersion

	//IdentifierRejected 0x02 Connection Refused, identifier rejected
	IdentifierRejected

	//ServiceUnavailable 0x03 Connection Refused, Server unavai
	ServiceUnavailable

	//WrongUserNameOrPass 0x04 Connection Refused, bad user name or password
	WrongUserNameOrPass

	//NotAuthorized 0x04 Connection Refused, bad user name or password
	NotAuthorized

	// Reserved for future use
	Reserved
)

//ConnAckMessage connection ack
type ConnAckMessage struct {
	FixedHeader
	flag    byte
	retCode byte
}

//IsSessionPresent return whether the session is existing
func (ack *ConnAckMessage) IsSessionPresent() bool {
	return ack.flag&0x1 == 0x1
}

func (ack *ConnAckMessage) SetSessionPresent(v bool) {
	if v {
		ack.flag |= 0x1
	} else {
		ack.flag &= 0xfe
	}
}

func (ack *ConnAckMessage) MessageType() byte {
	return CONNACK
}

func (ack *ConnAckMessage) MessageLen() int {
	bodyLen := 2
	ack.SetRemainLen(uint32(bodyLen))
	msgLen := bodyLen + ack.headerLen()
	return msgLen
}

func (ack *ConnAckMessage) Encode(buf []byte) (int, error) {

	msgLen := ack.MessageLen()
	if len(buf) < msgLen {
		return 0, errors.New("wrong buffer size ")
	}

	index := 0
	l, err := ack.encodeHeader(buf[index:])
	if err != nil {
		return 0, err
	}

	index += l
	l, err = ack.encodeMsg(buf[index:])
	if err != nil {
		return index, err
	}
	index += l
	return index, nil
}

func (ack *ConnAckMessage) encodeMsg(buf []byte) (int, error) {
	buf[0] = ack.flag
	buf[1] = ack.retCode
	return 2, nil
}

func (ack *ConnAckMessage) Decode(buf []byte) (int, error) {
	index := 0
	l, err := ack.decodeHeader(buf)
	if err != nil {
		return 0, err
	}
	index += l
	if len(buf[index:]) < int(ack.remainLen) {
		return index, errors.New("Invalid message: not have enough space to parse")
	}
	ack.flag = buf[index]
	index += 1
	ack.retCode = buf[index]
	return index, nil
}
func (ack *ConnAckMessage) Verify() error {
	return nil
}
