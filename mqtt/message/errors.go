package message

import "errors"

var (
	// ErrorInvalidTopic for invalid message topic
	ErrorInvalidTopic = errors.New("Invalid topic name pattern ")
)
