package message

import (
	"encoding/binary"
	"errors"
)

func parseVariableLen(buf []byte) (int, int, error) {
	multiplier := 1
	value := 0
	index := 0
	for _, b := range buf[:] {
		index++
		value += int(b&127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			return 0, 0, errors.New("malformed remain length")
		}
		if b&128 != 0 {
			break
		}
	}
	return value, index, nil
}

func readVint(buf []byte) (uint32, int, error) {
	v, l := binary.Uvarint(buf)
	return uint32(v), l, nil
	/*var multiplier uint32 = 1
	var value uint32
	index := 0
	for _, b := range buf[:] {
		index++
		value += uint32(b&127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			return 0, 0, errors.New("malformed remain length")
		}
		if b&128 == 0 {
			break
		}
	}
	return value, index, nil*/
}

func writeVint(buf []byte, v uint32) int {

	n := binary.PutUvarint(buf[:], uint64(v))
	return n

	/*
		i := 0
		for {
			encodeByte := v % 128
			v = v / 128
			if v > 0 {
				encodeByte = encodeByte | 0x80
			}
			buf[i] = byte(encodeByte)
			i++
			if v <= 0 {
				break
			}
		}
		return i*/
}

func vintLen(v uint32) int {
	i := 0
	for {
		v = v / 128
		i++
		if v <= 0 {
			break
		}
	}
	return i
}

func readLPBytes(msg []byte) ([]byte, int, error) {
	length := binary.BigEndian.Uint16(msg[:2])
	if len(msg) < int(length+2) {
		return nil, 0, errors.New("buffer not contain enough bytes to decode")
	}

	return msg[2 : 2+length], int(length + 2), nil
}

func writeLPBytes(dest []byte, src []byte) (int, error) {
	msgLen := len(src) + 2
	if len(dest) < msgLen {
		return 0, errors.New("not enough buffer to write")
	}

	binary.BigEndian.PutUint16(dest[:2], uint16(len(src)))
	copy(dest[2:], src)
	return msgLen, nil
}

func readUint16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf[:2])

}

func writeUint16(buf []byte, v uint16) {
	binary.BigEndian.PutUint16(buf, v)
}
