package events

import (
	"bytes"
	"time"
)

type User struct {
	Id    int64
	Login string
}

type Message struct {
	Id          int64
	SenderId    int64
	MessageType byte
	Message     string
	SendTime    time.Time
	ReadTime    time.Time
}

type Chat struct {
	ChatId   int64
	ChatName string
	Users    []User
	Messages []Message
}

type ClientEvent interface {
	Serialize() *bytes.Buffer
}
