package helper

import (
	"crypto/rsa"
	"encoding/gob"
	"net"
)

type Message struct {
	Message []byte
	Command string
	Key     *rsa.PublicKey
	From    string
}

func EncodeMessage(message Message, c net.Conn) {
	encoder := gob.NewEncoder(c)
	encoder.Encode(message)
}

func DecodeMessage(c net.Conn) Message {
	dec := gob.NewDecoder(c)
	message := &Message{}
	dec.Decode(message)
	return *message
}

func CreateMessage(content []byte, command string) *Message {
	message := new(Message)
	message.Command = command
	message.Message = content
	return message
}
