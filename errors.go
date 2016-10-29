package main

import (
	"errors"
)

var (
	errDisconnect = errors.New("Disconnect")
	errMsgFormat  = errors.New("WrongMsgFormat")
	errMsgSize    = errors.New("WrongMsgSize")
)
