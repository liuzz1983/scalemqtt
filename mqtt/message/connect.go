package message

import "errors"

// Connection message Variable header:
// Protocol Name 1-6
// Protocol Level: 7,
//       3.1.1 of the protocol is 4 (0x04), The Server MUST respond to the CONNECT Packet with a CONNACK return code 0x01 (unacceptable protocol level) and then disconnect the Client if the Protocol Level is not supported by the Server
// Connect Flags: 8
//       User Name Flag: 7
//       Password Flag: 6
//       Will Retain: 5
//       Will QoS: 4-3
//       Will Flag: 2
//       Clean Session: 1
//       Reserved :0
// Keep Alive:8-9
//       Keep Alive MSB: 9
//       Keep Alive LSB: 10
// Protocol Name: mqtt

// ConnMessage struct for connect message
type ConnMessage struct {
	FixedHeader
	ProtoName   []byte
	ProtoLevel  byte
	ConnectFlag byte
	KeepAlive   uint16

	clientID    []byte
	WillTopic   []byte
	WillMessage []byte
	UserName    []byte
	PassWord    []byte
}

// Encode encode connect message into network bytes
func (msg *ConnMessage) Encode(dst []byte) (int, error) {
	n, index := 0, 0
	totalLen := msg.headerLen() + msg.msgLen()
	if len(dst) < totalLen {
		return 0, errors.New("not enough space to store message")
	}

	n, err := msg.encodeHeader(dst[index:])
	if err != nil {
		return index, err
	}
	index += n

	n, err = msg.encodeMsg(dst[index:])
	if err != nil {
		return index, err
	}
	index += n

	return index, nil

}

func (msg *ConnMessage) encodeMsg(dest []byte) (int, error) {
	n, index := 0, 0
	msgLen := msg.msgLen()
	if len(dest) < msgLen {
		return 0, errors.New("not enough space to contain the msg")
	}

	n, err := writeLPBytes(dest[index:], msg.ProtoName)
	if err != nil {
		return index, err
	}
	index += n

	dest[index] = msg.ProtoLevel
	index++

	dest[index] = msg.ConnectFlag
	index++

	writeUint16(dest[index:], msg.KeepAlive)
	index += 2

	//write client id
	n, err = writeLPBytes(dest[index:], msg.clientID)
	if err != nil {
		return index, err
	}
	index += n

	if msg.IsWill() {
		n, err := writeLPBytes(dest[index:], msg.WillTopic)
		if err != nil {
			return index, err
		}
		index += n

		n, err = writeLPBytes(dest[index:], msg.WillMessage)
		if err != nil {
			return index, err
		}
		index += n
	}

	if msg.IsUserFlag() {
		n, err := writeLPBytes(dest[index:], msg.UserName)
		if err != nil {
			return index, err
		}
		index += n
	}

	if msg.IsPasswordFlag() {
		n, err = writeLPBytes(dest[index:], msg.PassWord)
		if err != nil {
			return index, err
		}
		index += n
	}

	return index, nil
}

// msgLen message length contain:
// variable header fixed 10
// These fields, if present, MUST appear in the order Client Identifier, Will Topic, Will Message, User Name, Password [MQTT-3.1.3-1].
func (msg *ConnMessage) msgLen() int {

	// msg length for fixed part
	msgLen := 10

	// client id
	msgLen += 2
	if msg.clientID != nil {
		msgLen += len(msg.clientID)
	}

	//will topic
	if msg.IsWill() {
		msgLen += 2
		msgLen += len(msg.WillTopic)

		msgLen += 2
		msgLen += len(msg.WillMessage)
	}

	if msg.IsUserFlag() {
		msgLen += 2
		msgLen += len(msg.UserName)
	}

	if msg.IsPasswordFlag() {
		msgLen += 2
		msgLen += len(msg.PassWord)
	}

	return msgLen
}

// Verify verify message
//
func (msg *ConnMessage) Verify() error {
	return errors.New("not implement")
}

// Version ConnectMessage version
func (msg *ConnMessage) Version() byte {
	return msg.ProtoLevel
}

// Decode decode connect message
func (msg *ConnMessage) Decode(buf []byte) (int, error) {

	index, err := msg.decodeHeader(buf)
	if err != nil {
		return 0, err
	}

	// begin to parse remain length
	remainMsg := buf[index:]
	// connection msg length is 9 byte
	if len(remainMsg) != 9 {
		return index, errors.New("wrong msg buf length")
	}

	//fmt.Println(len)
	return index, nil
}

// decodeMsg decode the connection message
// we have verify
func (msg *ConnMessage) decodeMsg(buf []byte) (int, error) {
	var err error
	index, n := 0, 0
	msg.ProtoName, n, err = readLPBytes(buf[index:])
	if err != nil {
		return n, err
	}
	index += n

	msg.ProtoLevel = buf[index]
	index++

	msg.ConnectFlag = buf[index]
	index++

	msg.KeepAlive = readUint16(buf[index:])
	index += 2

	msg.clientID, n, err = readLPBytes(buf[index:])
	if err != nil {
		return index, err
	}
	index += n

	if msg.IsWill() {
		msg.WillTopic, n, err = readLPBytes(buf[index:])
		if err != nil {
			return index, err
		}
		index += n

		msg.WillMessage, n, err = readLPBytes(buf[index:])
		if err != nil {
			return index, err
		}
		index += n
	}

	if msg.IsUserFlag() {
		msg.UserName, n, err = readLPBytes(buf[index:])
		if err != nil {
			return index, err
		}
		index += n
	}

	if msg.IsPasswordFlag() {
		msg.PassWord, n, err = readLPBytes(buf[index:])
		if err != nil {
			return index, err
		}
		index += n
	}
	return index, nil
}

// IsUserFlag whether the user flag is set
func (msg *ConnMessage) IsUserFlag() bool {
	return msg.ConnectFlag&0x80 == 0x80
}

// SetUserFlag set the user name flag
func (msg *ConnMessage) SetUserFlag(v bool) {
	if v {
		msg.ConnectFlag |= 0x80
	} else {
		msg.ConnectFlag &= 0x7f
	}
}

// IsPasswordFlag whether the password flag is set
func (msg *ConnMessage) IsPasswordFlag() bool {
	return msg.ConnectFlag&0x40 == 0x40
}

// SetPasswordFlag set password flag
func (msg *ConnMessage) SetPasswordFlag(v bool) {
	if v {
		msg.ConnectFlag |= 0x40
	} else {
		msg.ConnectFlag &= 0xbf
	}
}

// IsWillRetain wether the will retain flag is set
// will retain is set to 1 indicates that , if the connect request is accept,a will message must be stored in the server and associated
// with the network connecton
func (msg *ConnMessage) IsWillRetain() bool {
	return msg.ConnectFlag&0x20 == 0x20
}

// SetWillRetain set the will flag
func (msg *ConnMessage) SetWillRetain(v bool) {
	if v {
		msg.ConnectFlag |= 0x20
	} else {
		msg.ConnectFlag &= 0xdf
	}
}

// Qos get the message qos level
func (msg *ConnMessage) Qos() byte {
	return (msg.ConnectFlag & 0x18) >> 3
}

// SetQos set message qos level
//
func (msg *ConnMessage) SetQos(qos byte) error {
	if qos < QosAtMostOnce || qos > QosExactlyOnce {
		return errors.New("Invalid Qos Level")
	}
	msg.ConnectFlag |= qos<<3 | (msg.ConnectFlag & 0xe7)
	return nil
}

// IsWill judge whether the will flag is set
func (msg *ConnMessage) IsWill() bool {
	return (msg.ConnectFlag>>2)&0x1 == 1
}

// SetWill set connect will flag
func (msg *ConnMessage) SetWill(v bool) {
	if v {
		msg.ConnectFlag |= 0x4
	} else {
		msg.ConnectFlag &= 0xfb
	}
}

// IsCleanSession whether the clean session flag is set
func (msg *ConnMessage) IsCleanSession() bool {
	return (msg.ConnectFlag>>1)&0x1 == 1
}

// SetCleanSession set the clean session bit to identify whether clean the session state
func (msg *ConnMessage) SetCleanSession(v bool) {
	if v {
		msg.ConnectFlag |= 0x02
	} else {
		msg.ConnectFlag &= 0xfd
	}
}

// IsValidClientID verify the client id format
// The ClientId MUST be a UTF-8 encoded string as defined in Section 1.5.3 [MQTT-3.1.3-4].
func (msg *ConnMessage) IsValidClientID(clientId []byte) bool {

	if msg.Version() <= 0x3 {
		return true
	}

	return ClientIdPattern.Match(clientId)
}

// SetClientID set client and verify whether the clientid is right format
func (msg *ConnMessage) SetClientID(clientId []byte) error {

	if clientId == nil {
		return errors.New("wrong clientId")
	}
	if len(clientId) > 0 && !msg.IsValidClientID(clientId) {
		return errors.New("error inject client id")
	}

	if msg.clientID == nil {
		msg.clientID = make([]byte, len(clientId))
	}
	copy(msg.clientID, clientId)
	return nil
}
