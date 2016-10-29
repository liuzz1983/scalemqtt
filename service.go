package main

import (
	"errors"
	"github.com/surgemq/message"
	"net"
	"fmt"
	"sync/atomic"
)

var gsvcid uint64

type Service struct {

	// incremented for every new service.
	id uint64

	session *Session
	conn    net.Conn

	quit      chan struct{}
	parseChan chan []byte
	msgChan   chan message.Message
}

// 是否有必要将消息处理分为几个channel， 这样做有什么好处?
func NewService(session *Session, conn net.Conn) (service *Service) {
	return &Service{
		id:        atomic.AddUint64(&gsvcid, 1),
		session:   session,
		conn:      conn,
		quit:      make(chan struct{}),
		parseChan: make(chan []byte),
		msgChan:   make(chan message.Message),
	}

}

func (this *Service) close() {
	close(this.quit)
}

func (this *Service) processRecv() error {
	for {
		// for quit
		select {
		case <-this.quit:
			return nil
		default:
		}

		//TODO  if error, we need? close the channel
		mesageBytes, err := readMessage(this.conn)
		if err != nil {
			return nil
		}

		this.parseChan <- mesageBytes
	}
}

func (this *Service) loopMsg() error {

	for {
		select {
		case <-this.quit:
			return nil
		case msgBytes := <-this.parseChan:
			//
			msg, err := this.parseMsg(msgBytes)
			if err != nil {
				return err
			}
			this.msgChan <- msg
		}
	}
	return nil
}

func (this *Service) parseMsg(msgBytes []byte) (message.Message, error) {
	if msgBytes == nil || len(msgBytes) == 0 {
		return nil, errors.New("wrong msg")
	}

	// Extract the type from the first byte
	t := message.MessageType(msgBytes[0] >> 4)

	// Create a new message
	msg, err := t.New()
	if err != nil {
		return nil, err
	}

	// Decode it from the bufio.Reader
	_, err = msg.Decode(msgBytes)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (this *Service) processMsg(msg message.Message) error {
	switch ins := msg.(type) {
	case *message.PublishMessage:
		this.processPublish(ins)
	default:
		return fmt.Errorf("(%s) invalid message type %s.", this.cid(), msg.Name())
	}
	return nil
}

func (this *Service) cid() string {
	return fmt.Sprintf("%d", this.id)
}

//
func (this *Service) processPublish(msg *message.PublishMessage) error {
	return nil
}
