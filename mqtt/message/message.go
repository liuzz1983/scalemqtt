package message

import "errors"

// Message interface for basic message operation
type Message interface {
	MessageType() byte
	Encode([]byte) (int, error)
	Decode([]byte) (int, error)
	Verify() error
}

//MessageType create new message according to message type
func MessageType(t byte) (Message, error) {
	switch t {
	case CONNECT:
		return &ConnMessage{}, nil
	case CONNACK:
		return &ConnAckMessage{}, nil
	default:
		return nil, errors.New("invalid message type")
	}
}

// NewMessage build new message from bytes
func NewMessage(b []byte) (Message, error) {
	if b == nil || len(b) < 1 {
		return nil, errors.New("wrong message format")
	}
	msg, err := MessageType(b[0] >> 4)
	if err != nil {
		return nil, err
	}

	_, err = msg.Decode(b)
	if err != nil {
		return nil, err
	}

	// verify message format
	err = msg.Verify()
	return msg, err

}
