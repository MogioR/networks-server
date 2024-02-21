package events

import (
	"bytes"
	"encoding/binary"
	"time"
)

type NewMessageEvent struct {
	MessageId int64
	ChatId    int64
	UserId    int64
	IsFile    bool
	Message   string
	SendTime  time.Time
	ReadTime  time.Time
}

func (c *NewMessageEvent) Serialize() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, int8(-4)); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, c.MessageId); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, c.ChatId); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, c.UserId); err != nil {
		return nil, err
	}

	var isFileByte byte
	if c.IsFile {
		isFileByte = 1
	}
	if err := binary.Write(buf, binary.BigEndian, isFileByte); err != nil {
		return nil, err
	}

	messageLength := uint16(len(c.Message))
	if err := binary.Write(buf, binary.BigEndian, messageLength); err != nil {
		return nil, err
	}

	if _, err := buf.WriteString(c.Message); err != nil {
		return nil, err
	}

	postedAtUnix := c.SendTime.Unix()
	if err := binary.Write(buf, binary.BigEndian, postedAtUnix); err != nil {
		return nil, err
	}
	readedAtUnix := c.ReadTime.Unix()
	if err := binary.Write(buf, binary.BigEndian, readedAtUnix); err != nil {
		return nil, err
	}

	return buf, nil
}
