package mqtt

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/surgemq/message"
)

// Server basic structure
type Server struct {
	serviceId int64
	address   string

	connectTimeout time.Duration

	authMgr  Authentication
	sessMgr  *SessionManager
	topicMgr *TopicsManager

	quit chan struct{}
}

// NewServer create new server
// listen on socket to receive message
func NewServer(config *ServerConfig) (*Server, error) {
	server := &Server{
		address:        config.Address,
		connectTimeout: time.Duration(config.Timeout),

		sessMgr:  NewSessionManager(),
		topicMgr: NewTopicManager(),
		authMgr:  &NullAuth{},

		quit: make(chan struct{}, 1),
	}
	return server, nil
}

// Listen listen service
func (serv *Server) Listen() error {
	ln, err := net.Listen("tcp", serv.address)
	if err != nil {
		return err
	}
	defer ln.Close()

	serv.quit = make(chan struct{})
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := ln.Accept()
		if err != nil {

			// for quit
			select {
			case <-serv.quit:
				return nil
			default:
			}

			// Borrowed from go1.3.3/src/pkg/net/http/server.go:1699
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go serv.handleConnection(conn)
	}
}

func (serv *Server) Close() {
	serv.quit <- struct{}{}
}

// 建立超时设置？ 如何应对用户连接却不发送conn消息的情况
// 处理链接的过程：
// 1. 读取connection消息到buffer
// 2. 解析ConnectionMessage消息并验证消息格式,如果失败返回失败原因
// 3. 验证消息账户
// 4. 通知客户端，成功接收消息
// 5. 获取session， 如果没有则创建
// parse message
func (serv *Server) handleConnection(conn net.Conn) error {

	// ?? how to deal with this
	connTimeout := time.Now().Add(time.Second * serv.connectTimeout)
	conn.SetDeadline(connTimeout)

	// read message
	buf, err := ReadMessage(conn)
	if err != nil {
		return err
	}

	// parse connection message,has validated the msg
	resp := message.NewConnackMessage()

	// will return the following errors
	//0x00 Connection Accepted
	//0x01 Connection Refused, unacceptable protocol version
	//0x02 Connection Refused, identifier rejected
	//0x04 Connection Refused, bad user name or password
	//0x05 Connection Refused, not authorized
	req, err := serv.parseConnMsg(buf)
	if err != nil {
		if cerr, ok := err.(message.ConnackCode); ok {
			//glog.Debugf("request   message: %s\nresponse message: %s\nerror           : %v", mreq, resp, err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			WriteMessage(resp, conn)
		}
		return err
	}

	//TODO, add auth process logic
	//auth msg
	if !serv.authMgr.Auth(string(req.Username()), string(req.Password())) {
		resp.SetReturnCode(message.ErrNotAuthorized)
		WriteMessage(resp, conn)
		return errors.New("user is not authorized ")
	}

	// 通知client，成功接收消息
	WriteMessage(resp, conn)

	// TODO ?how to deal with this, when sesson get wrong, we should return id?
	// write ack message
	sess, err := serv.GetSession(req, resp)
	if err != nil {
		conn.Close()
		return err
	}

	// 递增serverid
	atomic.AddInt64(&serv.serviceId, 1)

	// add into service loop
	service := NewService(serv.serviceId, sess, conn, req, serv, serv.topicMgr)
	service.Start()

	return nil
}

// If CleanSession is set to 0, the server MUST resume communications with the
// client based on state from the current session, as identified by the client
// identifier. If there is no session associated with the client identifier the
// server must create a new session.
//
// If CleanSession is set to 1, the client and server must discard any previous
// session and start a new one. This session lasts as long as the network c
// onnection. State data associated with this session must not be reused in any
// subsequent session.
func (serv *Server) GetSession(req *message.ConnectMessage, resp *message.ConnackMessage) (*Session, error) {

	var err error

	// Check to see if the client supplied an ID, if not, generate one and set
	// clean session.
	// TODO

	if len(req.ClientId()) == 0 {
		req.SetClientId([]byte(fmt.Sprintf("internalclient %d", serv.serviceId)))
		req.SetCleanSession(true)
	}

	cid := string(req.ClientId())

	var session *Session

	// If CleanSession is NOT set, check the session store for existing session.
	// If found, return it.
	if !req.CleanSession() {
		if session, err = serv.sessMgr.Get(cid); err == nil {
			resp.SetSessionPresent(true)

			if err := session.Update(req); err != nil {
				return nil, err
			}
		}
	}

	// If CleanSession, or no existing session found, then create a new one
	if session == nil {
		if session, err = serv.sessMgr.New(cid); err != nil {
			return nil, err
		}

		resp.SetSessionPresent(false)
		if err := session.Init(req); err != nil {
			return nil, err
		}
	}

	return session, nil

}

func (serv *Server) parseConnMsg(buf []byte) (*message.ConnectMessage, error) {

	connMessage := message.NewConnectMessage()
	_, err := connMessage.Decode(buf)
	if err != nil {
		return nil, err
	}

	return connMessage, nil
}
