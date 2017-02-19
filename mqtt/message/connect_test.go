package message

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

//TestConnectMessage
func TestConnectMessageFlag(t *testing.T) {
	conn := &ConnMessage{}

	conn.SetPasswordFlag(true)
	assert.Equal(t, conn.IsPasswordFlag(), true, "password flag should be true")
	conn.SetPasswordFlag(false)
	assert.Equal(t, conn.IsPasswordFlag(), false, "password flag should be false")

	conn.SetUserFlag(true)
	assert.Equal(t, conn.IsUserFlag(), true, "user name flag should true")
	conn.SetUserFlag(false)
	assert.Equal(t, conn.IsUserFlag(), false, "user flag should false ")

	conn.SetWill(true)
	assert.Equal(t, conn.IsWill(), true, "will flag should be true")
	conn.SetWill(false)
	assert.Equal(t, conn.IsWill(), false, "will flag should be false ")

	err := conn.SetQos(QosAtMostOnce)
	assert.NoError(t, err, "set Qos should not return nil ")
	assert.Equal(t, conn.Qos(), QosAtMostOnce, "qos should be QosAtMostOnce")

	conn.SetCleanSession(true)
	assert.Equal(t, conn.IsCleanSession(), true, "clean session should be true")
	conn.SetCleanSession(false)
	assert.Equal(t, conn.IsCleanSession(), false, "clean session should be false ")

}

func TestConnectMessageEncoding(t *testing.T) {
	conn := NewConnMessage()
	clientID := []byte("0234448888333")
	conn.SetClientID(clientID)
	conn.SetUser([]byte("liuzhenzhong"), []byte("12344"))

	totalLen := conn.headerLen() + conn.msgLen()
	buf := make([]byte, totalLen)

	encodeLen, err := conn.Encode(buf)
	assert.NoError(t, err, "should not have error in encoding")
	fmt.Printf("%v \n", buf)

	conn2 := &ConnMessage{}
	decodeLen, err := conn2.Decode(buf)
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, encodeLen, decodeLen, "message should be equal")
}
