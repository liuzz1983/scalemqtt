package mqtt

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/surgemq/message"
)

type netReader interface {
	io.Reader
	SetReadDeadline(t time.Time) error
}

type timeoutReader struct {
	d    time.Duration
	conn netReader
}

func (r timeoutReader) Read(b []byte) (int, error) {
	if err := r.conn.SetReadDeadline(time.Now().Add(r.d)); err != nil {
		return 0, err
	}
	return r.conn.Read(b)
}

func WriteMessageWithTimeout(msg message.Message, conn net.Conn, timeout time.Duration) (int, error) {

	// set timeout for writing message
	if timeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return 0, err
		}
	}
	return WriteMessage(msg, conn)
}

func WriteMessage(msg message.Message, writer io.Writer) (int, error) {

	length := msg.Len()
	buf := make([]byte, length)

	msgLen, err := msg.Encode(buf[:])
	if err != nil {
		return 0, nil
	}

	if msgLen != length {
		return 0, ErrMsgSize
	}

	var totalLen = 0
	for {
		n, err := writer.Write(buf[totalLen:])
		if err != nil {

			// judge whether this is network temporary error
			// if netError, ok := err.(net.Error); ok && netError.Temporary() {
			//	continue
			//}
			// can not use this, because timeout is also temporary error
			// TODO, how to return the totalLen
			return totalLen, err
		}

		//TODO,need investigate this condition
		if n == 0 {
			continue
		}

		totalLen += n
		if totalLen >= len(buf) {
			return totalLen, nil
		}
	}
}

//this is wrong, because timeout is in the range
func IsNetworkTemporyError(err error) bool {
	if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
		return true
	}
	return false
}

func IsTimeoutError(err error) bool {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}

func ReadMessageWithTimeout(conn net.Conn, timeout time.Duration) ([]byte, error) {
	reader := &timeoutReader{
		conn: conn,
		d:    time.Duration(timeout) * time.Millisecond,
	}

	return ReadMessage(reader)
}

func ReadMessage(conn io.Reader) ([]byte, error) {
	// the message buffer
	var buf []byte
	// tmp buffer to read a single byte
	var b []byte = make([]byte, 1)
	// total bytes read
	var l int = 0

	for {

		n, err := conn.Read(b[:])
		if err != nil {
			//TODO can not use this code, maybe we should seperate temporary and timeout
			//if IsNetworkTemporyError(err) {
			//	break
			//}
			return nil, err
		}

		//TODO how to deal with this logic
		if n == 0 {
			continue
		}

		buf = append(buf, b...)
		l += n

		// Check the remlen byte (1+) to see if the continuation bit is set. If so,
		// increment cnt and continue reading. Otherwise break.
		// 1 0 (0x00) 127 (0x7F)
		// 2 128 (0x80, 0x01) 16383 (0xFF, 0x7F)
		// 3 16384 (0x80, 0x80, 0x01) 2097151 (0xFF, 0xFF, 0x7F)
		// 4 2097152 (0x80, 0x80, 0x80, 0x01) 268435455 (0xFF, 0xFF, 0xFF, 0x7F)
		if l > 1 && b[0] < 0x80 {
			break
		}

		//control the msg length
		if len(buf) >= 5 {
			return nil, ErrMsgFormat
		}

	}

	// Get the remaining length of the message
	remlen, _ := binary.Uvarint(buf[1:])
	buf = append(buf, make([]byte, remlen)...)

	// read the remaining message
	for l < len(buf) {
		n, err := conn.Read(buf[l:])
		if err != nil {
			return nil, err
		}
		l += n
	}

	return buf, nil
}
