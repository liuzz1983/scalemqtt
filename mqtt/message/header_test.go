package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {

	type Case struct {
		t  byte
		l  uint32
		rl int
		d  []byte
	}

	cases := []Case{
		{
			CONNECT,
			3,
			2,
			[]byte{0x10, 0x3},
		},

		{
			CONNACK,
			129,
			3,
			[]byte{0x20, 0x81, 0x01},
		},
	}

	for _, c := range cases {
		header := &FixedHeader{}
		header.SetMessageType(c.t)
		header.SetRemainLen(c.l)

		dest := make([]byte, c.rl)
		l, err := header.encodeHeader(dest)

		assert.Equal(t, l, c.rl, "header size shoul be 2")
		assert.NoError(t, err, "encode header shoul not be nil ")
		assert.Equal(t, dest, c.d, "encode header should equal")

		destHeader := &FixedHeader{}
		l, err = destHeader.decodeHeader(dest)
		assert.Equal(t, destHeader.MessageType(), c.t, "message type should be equal")
		assert.NoError(t, err, "decode should not return error ")
		assert.Equal(t, destHeader.remainLen, c.l, "remain length should be equal")
		assert.Equal(t, l, c.rl, "message length should be equal")

	}

}
