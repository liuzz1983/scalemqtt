package main

import (
	"bytes"
	"fmt"
	"github.com/liuzz1983/scalemqtt/mqtt"
	"github.com/surgemq/message"
	"io"
	"net"
	"sync/atomic"
	"time"
)

func SendMsg(conn io.Writer, msg message.Message) error {
	buf := make([]byte, msg.Len())
	n, err := msg.Encode(buf)
	if err != nil {
		return err
	}

	_, err = io.CopyN(conn, bytes.NewReader(buf), int64(n))
	if err != nil {
		fmt.Printf("error in copy %v", err)
		return err
	}
	return nil
}

type Client struct {
	conn  net.Conn
	count int32
}

func (client *Client) init() error {

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Printf("error in dial sever %v", err)
		return err
	}
	client.conn = conn
	return nil
}

func (client *Client) buildConnection() error {

	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
		}
	}()

	// Create a new CONNECT message
	msg := message.NewConnectMessage()

	// Set the appropriate parameters
	msg.SetWillQos(1)
	msg.SetVersion(4)
	msg.SetCleanSession(true)
	msg.SetClientId([]byte("surgemq"))
	msg.SetKeepAlive(10)
	msg.SetWillTopic([]byte("will"))
	msg.SetWillMessage([]byte("send me home"))
	msg.SetUsername([]byte("surgemq"))
	msg.SetPassword([]byte("verysecret"))

	// Encode the message and get the io.Reader
	buf := make([]byte, msg.Len())
	n, err := msg.Encode(buf)
	if err != nil {
		fmt.Printf("error,msg%v\n", err)
		return err
	}
	// Write n bytes into the connection
	_, err = io.CopyN(client.conn, bytes.NewReader(buf), int64(n))
	if err != nil {
		fmt.Printf("error,%v\n", err)
		return err
	}

	// fmt.Printf("Sent %d bytes of %s message\n", m, msg.Name())

	value, err := mqtt.ReadMessage(client.conn)
	if err != nil {
		fmt.Printf("error in read value\n", err)
		return err
	}
	ackMsg := message.NewConnackMessage()

	_, err = ackMsg.Decode(value)
	if err != nil {
		fmt.Printf("error,%v\n", err)
		return err
	}

	//fmt.Printf("message %v\n", ackMsg)
	return nil

}

func (client *Client) subMsg() error {

	subMsg := message.NewSubscribeMessage()
	subMsg.AddTopic([]byte("/time"), byte(0))
	SendMsg(client.conn, subMsg)
	count := 0

	for {
		value, _ := mqtt.ReadMessage(client.conn)
		pubMsg := message.NewPublishMessage()
		_, err := pubMsg.Decode(value)
		if err != nil {
			fmt.Printf("%v, %v\n", pubMsg, err)
		}

		count += 1
		if count%1000 == 0 {
			fmt.Printf("%v\n", string(pubMsg.Payload()))
		}

		// time.Sleep(1 * time.Second)
	}
}

func (client *Client) pubMsg() error {

	pubMsg := message.NewPublishMessage()
	pubMsg.SetTopic([]byte("/time"))

	_ = atomic.AddInt32(&client.count, 1)
	pubMsg.SetPayload([]byte(fmt.Sprintf("hello world %v", client.count)))
	return SendMsg(client.conn, pubMsg)

}

func main() {
	for j := 0; j < 1000; j++ {
		c := &Client{}
		err := c.init()
		if err != nil {
			fmt.Printf("error in build connection %v", err)
			continue
		}
		c.buildConnection()
		go func() {
			c.subMsg()

		}()
	}

	for i := 0; i < 1000; i++ {
		c := &Client{}
		err := c.init()
		if err != nil {
			fmt.Printf("error in build connection %v", err)
			continue
		}
		c.buildConnection()
		go func() {
			for {
				err := c.pubMsg()
				if err != nil {
					break
				}
			}
		}()
	}

	for {
		time.Sleep(1 * time.Second)
	}

}
