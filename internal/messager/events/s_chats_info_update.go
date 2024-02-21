package events

import (
	"bytes"
	"encoding/binary"
)

type ChatEvent struct {
	Chats []*Chat
}

func (c *ChatEvent) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, int8(-2)); err != nil {
		return nil, err
	}

	// Сериализуем количество чатов
	numChats := int8(len(c.Chats))
	if err := binary.Write(buf, binary.BigEndian, numChats); err != nil {
		return nil, err
	}

	// Сериализуем каждый чат
	for _, chat := range c.Chats {
		serializedChat, err := chat.Serialize()
		if err != nil {
			return nil, err
		}
		buf.Write(serializedChat.Bytes())
	}

	return buf.Bytes(), nil
}

func (chat *Chat) Serialize() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	// Сериализуем Id чата
	if err := binary.Write(buf, binary.BigEndian, chat.ChatId); err != nil {
		return nil, err
	}

	// Сериализуем длину имени чата
	buf.WriteByte(byte(len(chat.ChatName)))

	// Сериализуем имя чата
	buf.WriteString(chat.ChatName)

	// Сериализуем количество пользователей в чате
	numUsers := byte(len(chat.Users))
	if err := binary.Write(buf, binary.BigEndian, numUsers); err != nil {
		return nil, err
	}

	// Сериализуем каждого пользователя
	for _, user := range chat.Users {
		if err := binary.Write(buf, binary.BigEndian, user.Id); err != nil {
			return nil, err
		}

		// Сериализуем длину логина пользователя
		buf.WriteByte(byte(len(user.Login)))

		// Сериализуем логин пользователя
		buf.WriteString(user.Login)
	}

	// Сериализуем количество сообщений в чате
	numMessages := byte(len(chat.Messages))
	if err := binary.Write(buf, binary.BigEndian, numMessages); err != nil {
		return nil, err
	}

	// Сериализуем каждое сообщение
	for _, message := range chat.Messages {
		serializedMessage, err := message.Serialize()
		if err != nil {
			return nil, err
		}
		buf.Write(serializedMessage.Bytes())
	}

	return buf, nil
}

func (message *Message) Serialize() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	// Сериализуем Id сообщения
	if err := binary.Write(buf, binary.BigEndian, message.Id); err != nil {
		return nil, err
	}

	// Сериализуем Id отправителя
	if err := binary.Write(buf, binary.BigEndian, message.SenderId); err != nil {
		return nil, err
	}

	// Сериализуем тип сообщения
	if err := binary.Write(buf, binary.BigEndian, message.MessageType); err != nil {
		return nil, err
	}

	// Сериализуем длину сообщения
	messageLength := uint16(len(message.Message))
	if err := binary.Write(buf, binary.BigEndian, messageLength); err != nil {
		return nil, err
	}

	// Сериализуем сообщение
	buf.WriteString(message.Message)

	// Сериализуем временные метки
	postedAtUnix := message.SendTime.Unix()
	if err := binary.Write(buf, binary.BigEndian, postedAtUnix); err != nil {
		return nil, err
	}
	readedAtUnix := message.ReadTime.Unix()
	if err := binary.Write(buf, binary.BigEndian, readedAtUnix); err != nil {
		return nil, err
	}

	return buf, nil
}
