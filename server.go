package main

import (
	"fmt"
	"github.com/surgemq/message"
	"net"
	"time"
)

type Server struct {
	address string
	authMgr Auth

	quit chan struct{}

	sessMgr *SessionManager
}

func NewServer(address string) (*Server, error) {
	server := &Server{
		address: address,
		quit:    make(chan struct{}, 1),
	}
	return server, nil
}

func (serv *Server) listen() error {
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

// parse message
func (serv *Server) handleConnection(conn net.Conn) error {

	// ?? how to deal with this
	connectTimeout := time.Duration(20)
	conn.SetDeadline(time.Now().Add(time.Second * connectTimeout))

	resp := message.NewConnackMessage()

	// parse connection message,has validated the msg
	_, err := serv.parseConnMsg(conn)

	if err != nil {
		if cerr, ok := err.(message.ConnackCode); ok {
			//glog.Debugf("request   message: %s\nresponse message: %s\nerror           : %v", mreq, resp, err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(resp, conn)
		}
		return err
	}

	//auth msg
	//0x00 Connection Accepted
	//0x01 Connection Refused, unacceptable protocol version
	//0x02 Connection Refused, identifier rejected
	//0x04 Connection Refused, bad user name or password
	//0x05 Connection Refused, not authorized
	writeMessage(resp, conn)

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
func (this *Server) getSession(svc *Service, req *message.ConnectMessage, resp *message.ConnackMessage) error {

	var err error

	// Check to see if the client supplied an ID, if not, generate one and set
	// clean session.
	if len(req.ClientId()) == 0 {
		req.SetClientId([]byte(fmt.Sprintf("internalclient%d", svc.id)))
		req.SetCleanSession(true)
	}

	cid := string(req.ClientId())

	// If CleanSession is NOT set, check the session store for existing session.
	// If found, return it.
	if !req.CleanSession() {
		if svc.session, err = this.sessMgr.Get(cid); err == nil {
			resp.SetSessionPresent(true)

			if err := svc.session.Update(req); err != nil {
				return err
			}
		}
	}

	// If CleanSession, or no existing session found, then create a new one
	if svc.session == nil {
		if svc.session, err = this.sessMgr.New(cid); err != nil {
			return err
		}

		resp.SetSessionPresent(false)

		if err := svc.session.Init(req); err != nil {
			return err
		}
	}

	return nil

}

func (serv *Server) parseConnMsg(conn net.Conn) (*message.ConnectMessage, error) {

	buf, err := readMessage(conn)
	if err != nil {
		return nil, err
	}

	connMessage := message.NewConnectMessage()
	_, err = connMessage.Decode(buf)
	if err != nil {
		return nil, err
	}

	return connMessage, nil
}

func main() {
	server, _ := NewServer(":8080")
	server.listen()
}
