package events

import (
	"bytes"
	"encoding/binary"
)

type GetFileFromChatEvent struct {
	ChatId    int64
	MessageId int64
}

func (e *GetFileFromChatEvent) Deserialize(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &e.ChatId); err != nil {
		return err
	}

	if err := binary.Read(buf, binary.BigEndian, &e.MessageId); err != nil {
		return err
	}

	return nil
}
