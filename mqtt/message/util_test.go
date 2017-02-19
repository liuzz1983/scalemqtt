package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLPWrite(t *testing.T) {
	src := []byte("time")
	dest := make([]byte, len(src)+2)

	n, err := writeLPBytes(dest, src)
	assert.NoError(t, err, "should not return err for write bytes")
	assert.Equal(t, n, 6, "msg length should be equal")
	assert.Equal(t, dest, []byte{0x00, 0x04, 't', 'i', 'm', 'e'}, "bytes should not be equal")

	v, n, err := readLPBytes(dest)
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, n, len(src)+2, "length should be equal ")
	assert.Equal(t, v, src, "content should be equal")

}

func TestVint(t *testing.T) {
	type Case struct {
		v uint32
		r []byte
	}

	cases := []Case{
		{
			2,
			[]byte{0x2},
		},

		{
			129,
			[]byte{0x81, 0x01},
		},
	}

	for _, c := range cases {
		v := uint32(c.v)
		buf := make([]byte, 4)
		l := writeVint(buf, v)
		assert.Equal(t, buf[:l], c.r, "sequence should be equal ")
		assert.Equal(t, l, len(c.r), "message length should be equal")

		r, l, _ := readVint(buf)
		assert.NotEqual(t, r, l, "length sould not be equal")
	}

}
