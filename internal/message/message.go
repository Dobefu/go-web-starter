package message

import "encoding/gob"

type MessageType int

const (
	MessageTypeError = iota
	MessageTypeWarning
	MessageTypeInfo
	MessageTypeSuccess
)

type Message struct {
	Type MessageType
	Body string
}

func init() {
	gob.Register(Message{})
}
