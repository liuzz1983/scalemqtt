package mqtt

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/surgemq/message"
)

var gsvcid uint64

// Service  主要的消息处理逻辑，读取消息并处理
// 增加metrics的统计功能
// TODO, 需要handle的几个问题：
// 1. 如何处理timeout
type Service struct {

	// incremented for every new service.
	id int64

	conn         net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration

	// need to process,this
	keepAlive uint16

	session *Session
	topics  *TopicsManager

	parseChan chan []byte
	msgChan   chan message.Message

	quit chan struct{}
}

// NewService 创建新的
// 是否有必要将消息处理分为几个channel， 这样做有什么好处?
func NewService(id int64, session *Session, conn net.Conn, connMsg *message.ConnectMessage,
	server *Server, topics *TopicsManager) (service *Service) {

	return &Service{
		id:           id,
		conn:         conn,
		writeTimeout: time.Duration(1) * time.Second,
		readTimeout:  time.Duration(1) * time.Second,

		keepAlive: connMsg.KeepAlive(),
		session:   session,

		parseChan: make(chan []byte),
		msgChan:   make(chan message.Message),
		topics:    topics,
		quit:      make(chan struct{}),
	}

}

func (service *Service) cid() string {
	return fmt.Sprintf("%d", service.id)
}

// Start 开始服务
// TODO how to deal with this situation
func (service *Service) Start() error {
	go service.loopReadMsg()
	go service.loopParseMsg()
	go service.loopProcessMsg()

	return nil
}

// Close 继续服务
func (service *Service) Close() {
	close(service.quit)
}

func (service *Service) readMessage() ([]byte, error) {

	reader := timeoutReader{
		d:    service.readTimeout * time.Second,
		conn: service.conn,
	}

	// 读取消息，在读取消息失败的情况下，需要关闭连接，并关闭service？
	mesageBytes, err := ReadMessage(reader)
	return mesageBytes, err
}

// TODO，如果写失败的情况下，需要判断是否是临时错误?,是否需要关闭通道
func (service *Service) writeMessage(msg message.Message) (int, error) {

	//if err := service.conn.SetWriteDeadline(time.Now().Add(service.writeTimeout * time.Second)); err != nil {
	//	return 0, err
	//}
	n, err := WriteMessage(msg, service.conn)
	if err != nil {
		glog.Errorf("error in write msg %v %v ", n, err)
		return n, err
	}
	return n, nil
}

func (service *Service) loopOne() error {

	for {
		// for quit
		select {
		case <-service.quit:
			return nil
		default:
		}

		msgBytes, err := service.readMessage()
		if err != nil {
			// TODO if error, we need? close the channel
			// 在不是timeout error的情况下，我们需要关闭连接和其他loop
			if !IsTimeoutError(err) {
				glog.Errorf("receive closed msg %v", err)
				service.conn.Close()

				// TODO 是否需要再考虑考虑
				// close other channel
				service.Close()
				return err
			}
			continue
		}

		msg, err := service.parseMsg(msgBytes)
		if err != nil {
			glog.Errorf("error in parse msg %v", err)
			return err
		}

		err = service.processMsg(msg)
		if err != nil {
			glog.Errorf("error in process message %v %v", msg, err)
			return err
		}

	}

}

func (service *Service) loopReadMsg() error {
	for {
		// for quit
		select {
		case <-service.quit:
			return nil
		default:
		}

		mesageBytes, err := service.readMessage()
		if err != nil {

			// TODO  if error, we need? close the channel
			// 在不是timeout error的情况下，我们需要关闭连接和其他loop
			if !IsTimeoutError(err) {
				fmt.Printf("receive closed msg %v \n", err)
				service.conn.Close()

				// close other channel
				service.Close()
				return nil
			}
			continue
		}

		service.parseChan <- mesageBytes
	}
}

func (service *Service) loopParseMsg() error {

	for {
		select {
		case <-service.quit:
			return nil
		case msgBytes := <-service.parseChan:
			//
			msg, err := service.parseMsg(msgBytes)
			if err != nil {
				fmt.Printf("error in parse msg %v\n", err)
				return err
			}
			service.msgChan <- msg
		}
	}
	return nil
}

func (service *Service) loopProcessMsg() error {
	for {
		select {
		case <-service.quit:
			return nil
		case msg := <-service.msgChan:
			err := service.processMsg(msg)
			if err != nil {
				fmt.Printf("error in process message %v %v\n", msg, err)
				return err
			}
		}
	}
	return nil
}

func (service *Service) parseMsg(msgBytes []byte) (message.Message, error) {
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

func (service *Service) processMsg(msg message.Message) error {

	switch ins := msg.(type) {
	case *message.PublishMessage:
		service.processPublish(ins)
	case *message.SubscribeMessage:
		service.processSubscribeMessage(ins)
	case *message.UnsubscribeMessage:
	default:
		return fmt.Errorf("(%v) invalid message type %v", service.cid(), msg.Name())
	}
	return nil
}

func (service *Service) processUnsubscribe(msg *message.UnsubscribeMessage) error {
	return nil
}

//
func (service *Service) processPublish(msg *message.PublishMessage) error {

	topic := string(msg.Topic())
	subs := service.topics.Find(topic)
	for _, sub := range subs {
		go sub.publish(msg)
	}
	return nil
}

// implement
func (service *Service) publish(msg *message.PublishMessage) error {

	n, err := service.writeMessage(msg)
	if err != nil {
		fmt.Printf("error in write msg %v %v ", n, err)
	}

	return err
}

func (service *Service) processSubscribeMessage(msg *message.SubscribeMessage) error {
	for _, topic := range msg.Topics() {
		service.topics.Register(string(topic), service.cid(), service)
	}
	return nil
}
