package mqtt

import (
	"strings"
	"sync"

	"github.com/surgemq/message"
)

// The topic level separator is used to introduce structure into the Topic Name. If present,
// it divides the Topic Name into multiple “topic levels”.
// A subscription’s Topic Filter can contain special wildcard characters, which allow you to subscribe to multiple topics at once.
// client publish topic ------ server
//				  find matched message, filter by topic filter
// client subscribe topic --- server, topic filter?
//							  how to deal with new incoming message
//
// the character must be specified either on its own or following a topic level seperator, in either case, it must be the last character specified
// by the topic filter.
// sport/tennis/player1/#  will receive message published by the following topic name
// sport/tennis/player1/
// sport/tennis/player1/ranking
// sport/tennis/player1/score/wimbledon

// sport/#” also matches the singular “sport”, since # includes the parent level.   “#” is valid and will receive every Application Message
// “sport/tennis/#” is valid
// “sport/tennis#” is not valid
// “sport/tennis/#/ranking” is not valid

// The single-level wildcard can be used at any level in the Topic Filter, including first and last levels.
//   “+” is valid
//  “+/tennis/#” is valid
//  “sport+” is not valid
//  “sport/+/player1” is valid
//  “/finance” matches “+/+” and “/+”, but not “+”

//now we need define a structure to handle the topic
// tree node1 / node2 / node3 / node4
//			  / node5
// 			  / node6 / node7
//TopicsManager manage topics
//filter
// we should build the filter tree or the topic name tree
// no doubt, we should build the filter tree
// manage topics
//

const (
	// MWC is the multi-level wildcard
	MWC = "#"

	// SWC is the single level wildcard
	SWC = "+"

	// SEP is the topic level separator
	SEP = "/"

	// SYS is the starting character of the system level topics
	SYS = "$"

	// Both wildcards
	_WC = "#+"
)

type Node struct {
	name     string
	parent   *Node
	children []*Node
}

type Tree struct {
	nodes []*Node
}

type Sub interface {
	publish(*message.PublishMessage) error
}

type TopicsManager struct {
	topicToSession map[string][]string
	sessionToSub   map[string]Sub
	lock           sync.Mutex
}

func NewTopicManager() *TopicsManager {
	return &TopicsManager{
		topicToSession: make(map[string][]string),
		sessionToSub:   make(map[string]Sub),
	}
}

func (manager *TopicsManager) Register(topic string, sessionId string, sub Sub) error {
	manager.lock.Lock()
	go manager.lock.Unlock()

	if sessions, ok := manager.topicToSession[topic]; ok {
		manager.topicToSession[topic] = append(sessions, sessionId)
	} else {
		manager.topicToSession[topic] = append(make([]string, 0, 10), sessionId)
	}

	manager.sessionToSub[sessionId] = sub
	return nil
}

func (manager *TopicsManager) Deregister(sessionId string) {
	manager.lock.Lock()
	go manager.lock.Unlock()

}

func (manager *TopicsManager) Find(topic string) []Sub {
	manager.lock.Lock()
	go manager.lock.Unlock()

	subs := make([]Sub, 0, 16)
	for k, vs := range manager.topicToSession {
		if manager.Match(topic, k) {
			for _, v := range vs {
				subs = append(subs, manager.sessionToSub[v])
			}
		}
	}
	return subs
}

func (manager *TopicsManager) Match(topic string, sub string) bool {

	if strings.Compare(topic, sub) == 0 {
		return true
	}
	return false
}
