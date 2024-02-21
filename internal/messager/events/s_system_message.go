package events

import (
	"bytes"
)

type SystemMessageEvent struct {
	Code    int8
	Message string
}

func (c *SystemMessageEvent) Serialize() *bytes.Buffer {

	buf := new(bytes.Buffer)

	id := int8(-1)
	buf.WriteByte(byte(id))
	buf.WriteByte(byte(c.Code))

	buf.WriteByte(byte(len(c.Message)))
	buf.WriteString(c.Message)

	return buf
}
