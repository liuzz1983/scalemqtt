package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/surgemq/message"
	"testing"
)

func TestReadConnectMessage(t *testing.T) {
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
	byteMsg := make([]byte, 200)
	_, err := msg.Encode(byteMsg[:])
	assert.Nil(t, err)

	reader := bytes.NewReader(byteMsg)

	buf, err := readMessage(reader)
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	msg2 := message.NewConnectMessage()
	msg2.Decode(buf)

	assert.Equal(t, msg2.WillQos(), uint8(1), "will qos should be 1")
	assert.Equal(t, msg2.Version(), uint8(4), "version should be 4")
	assert.Equal(t, msg2.CleanSession(), true, "clean session is set")
	assert.Equal(t, msg2.ClientId(), []byte("surgemq"), "client id should be surgemq")
	assert.Equal(t, msg2.Password(), []byte("verysecret"), "password should be equal")
	assert.Equal(t, msg2.Username(), []byte("surgemq"), "user should be equal")
	assert.Equal(t, msg2.WillMessage(), []byte("send me home"), "will message should be equal")
	assert.Equal(t, msg2.WillTopic(), []byte("will"), "topic msg shoudl be equal")
}
