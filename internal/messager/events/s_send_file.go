package events

import (
	"bytes"
	"encoding/binary"
)

type SendFileToChatEvent struct {
	FileName string
	FileSize int32
}

func (c *SendFileToChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	l := int8(-6)
	buf.WriteByte(byte(l))

	binary.Write(buf, binary.BigEndian, int16(len(c.FileName)))
	buf.WriteString(c.FileName)

	binary.Write(buf, binary.BigEndian, c.FileSize)

	return buf
}
