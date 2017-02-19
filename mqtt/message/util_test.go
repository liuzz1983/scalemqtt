package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
