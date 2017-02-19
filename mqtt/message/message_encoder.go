package message

type Buffer struct {
	buf []byte
	pos int
	l   int
	err error
}

//NewMessageBuffer create new message buffer
func NewMessageBuffer(buf []byte, l int) *Buffer {
	return &Buffer{
		buf: buf,
		l:   l,
		pos: 0,
	}
}

func (buf *Buffer) putUint16(v uint16) {
    
}
