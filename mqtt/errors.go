package mqtt

import (
	"errors"
)

var (
	ErrDisconnect = errors.New("Disconnect")
	ErrMsgFormat  = errors.New("WrongMsgFormat")
	ErrMsgSize    = errors.New("WrongMsgSize")
)
