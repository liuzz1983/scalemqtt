package main

import (
	"github.com/surgemq/message"
	"net"
)

type Session struct {
	conn net.Conn
}

func (this *Session) Init(msg *message.ConnectMessage) error {
	return nil
}

func (this *Session) Update(msg *message.ConnectMessage) error {
	return nil
}
