package events

import (
	"bytes"
	"encoding/binary"
)

type SendFileInitEvent struct {
	ChatId   int64
	FileName string
	FileSize int32
}

func (c *SendFileInitEvent) Deserialize(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &c.ChatId); err != nil {
		return err
	}
	var messageLength uint16
	if err := binary.Read(buf, binary.BigEndian, &messageLength); err != nil {
		return err
	}

	messageBytes := make([]byte, messageLength)
	if _, err := buf.Read(messageBytes); err != nil {
		return err
	}
	c.FileName = string(messageBytes)

	if err := binary.Read(buf, binary.BigEndian, &c.FileSize); err != nil {
		return err
	}

	return nil
}
