package storage

import "time"

type UserModel struct {
	Id    int64
	Login string
}

type MessageModel struct {
	Id       int64
	ChatId   int64
	UserId   int64
	IsFile   bool
	Message  string
	ReadedAt time.Time
	PostedAt time.Time
}

type ChatModel struct {
	Id          int64
	LastMessage time.Time
	Name        string
	Users       []UserModel
	Messages    []MessageModel
}
