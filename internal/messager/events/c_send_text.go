package events

import (
	"bytes"
	"encoding/binary"
)

type SendTextEvent struct {
	ChatId  int64
	Message string
}

func (e *SendTextEvent) Deserialize(buf *bytes.Buffer) error {

	if err := binary.Read(buf, binary.BigEndian, &e.ChatId); err != nil {
		return err
	}

	// Читаем длину сообщения
	messageLength, err := buf.ReadByte()
	if err != nil {
		return err
	}

	// Читаем само сообщение
	messageBytes := make([]byte, messageLength)
	if _, err := buf.Read(messageBytes); err != nil {
		return err
	}
	e.Message = string(messageBytes)

	return nil
}
