package messager

import (
	"bytes"
	"encoding/binary"
	"messager-server/internal/storage"
)

func (m *Messager) serializeChats(chats []storage.ChatModel) ([]byte, error) {
	var buf bytes.Buffer

	// Номер команды
	buf.WriteByte(1)

	// Количество чатов
	buf.WriteByte(byte(len(chats)))

	for _, chat := range chats {
		b, err := m.serializeChat(chat)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}

	return buf.Bytes(), nil
}

func (m *Messager) serializeMessage(message storage.MessageModel) ([]byte, error) {
	var buf bytes.Buffer

	// Тип события
	buf.WriteByte(2)

	// Id чата
	if err := binary.Write(&buf, binary.BigEndian, message.ChatId); err != nil {
		return nil, err
	}

	// Id сообщения
	if err := binary.Write(&buf, binary.BigEndian, message.Id); err != nil {
		return nil, err
	}

	// Id отправителя
	if err := binary.Write(&buf, binary.BigEndian, message.UserId); err != nil {
		return nil, err
	}

	// Тип сообщения
	var messageTypeByte byte
	if message.IsFile {
		messageTypeByte = 1
	}
	buf.WriteByte(messageTypeByte)

	// Длина сообщения
	messageLength := int16(len(message.Message))
	if err := binary.Write(&buf, binary.BigEndian, messageLength); err != nil {
		return nil, err
	}

	// Сообщение
	buf.WriteString(message.Message)

	// Время отправки
	postedAtUnix := message.PostedAt.Unix()
	if err := binary.Write(&buf, binary.BigEndian, postedAtUnix); err != nil {
		return nil, err
	}

	// Время прочтения
	readedAtUnix := message.ReadedAt.Unix()
	if err := binary.Write(&buf, binary.BigEndian, readedAtUnix); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Messager) serializeMessageReadAt(message storage.MessageModel) ([]byte, error) {
	var buf bytes.Buffer

	// Тип события
	buf.WriteByte(3)

	// Id чата
	if err := binary.Write(&buf, binary.BigEndian, message.ChatId); err != nil {
		return nil, err
	}

	// Id сообщения
	if err := binary.Write(&buf, binary.BigEndian, message.Id); err != nil {
		return nil, err
	}

	// Время первого прочтения
	firstReadUnix := message.ReadedAt.Unix()
	if err := binary.Write(&buf, binary.BigEndian, firstReadUnix); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *Messager) serializeChat(chat storage.ChatModel) ([]byte, error) {
	var buf bytes.Buffer

	// Чат айди
	if err := binary.Write(&buf, binary.BigEndian, chat.Id); err != nil {
		return nil, err
	}

	// Размер названия чата и само название
	buf.WriteByte(byte(len(chat.Name)))
	buf.WriteString(chat.Name)

	// Количество пользователей в чате
	buf.WriteByte(byte(len(chat.Users)))

	for _, user := range chat.Users {
		// Пользователь айди
		if err := binary.Write(&buf, binary.BigEndian, user.Id); err != nil {
			return nil, err
		}

		// Размер имени пользователя и само имя
		buf.WriteByte(byte(len(user.Login)))
		buf.WriteString(user.Login)
	}

	messages, err := m.storage.GetMessagesFromChat(chat.Id)
	if err != nil {
		return nil, err
	}

	// Количество сообщений в чате
	numMessages := int8(len(messages))
	if err := binary.Write(&buf, binary.BigEndian, numMessages); err != nil {
		return nil, err
	}

	// Сериализация сообщений для чата
	for _, msg := range messages {
		// Сообщение айди и Пользователь айди
		if err := binary.Write(&buf, binary.BigEndian, msg.Id); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.BigEndian, msg.UserId); err != nil {
			return nil, err
		}

		// Флаг наличия файла
		if msg.IsFile {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}

		// Длина сообщения и само сообщение
		msgLen := int16(len(msg.Message))
		if err := binary.Write(&buf, binary.BigEndian, msgLen); err != nil {
			return nil, err
		}
		if _, err := buf.WriteString(msg.Message); err != nil {
			return nil, err
		}

		// Дата публикации и прочтения сообщения
		if err := binary.Write(&buf, binary.BigEndian, msg.PostedAt.Unix()); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.BigEndian, msg.ReadedAt.Unix()); err != nil {
			return nil, err
		}
	}

	// Время отправки
	postedAtUnix := chat.LastMessage.Unix()
	if err := binary.Write(&buf, binary.BigEndian, postedAtUnix); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func serializeChangeChatMembers(chat storage.ChatModel) ([]byte, error) {
	var buf bytes.Buffer

	// Тип события
	buf.WriteByte(5)

	// Id чата
	if err := binary.Write(&buf, binary.BigEndian, chat.Id); err != nil {
		return nil, err
	}

	// Количество пользователей в чате
	buf.WriteByte(byte(len(chat.Users)))

	// Пользователи в чате
	for _, user := range chat.Users {
		// Id пользователя
		if err := binary.Write(&buf, binary.BigEndian, user.Id); err != nil {
			return nil, err
		}

		// Длинна логина пользователя
		buf.WriteByte(byte(len(user.Login)))

		// Логин пользователя
		buf.WriteString(user.Login)
	}

	return buf.Bytes(), nil
}

func deseralizeChatCreate(data []byte) (storage.ChatModel, error) {
	var chat storage.ChatModel

	buf := bytes.NewReader(data)

	// Пропускаем тип события
	if _, err := buf.ReadByte(); err != nil {
		return chat, err
	}

	// Чтение длины имени чата
	nameLen, err := buf.ReadByte()
	if err != nil {
		return chat, err
	}

	// Чтение названия чата
	nameBytes := make([]byte, nameLen)
	if _, err := buf.Read(nameBytes); err != nil {
		return chat, err
	}
	chat.Name = string(nameBytes)

	// Чтение количества пользователей в чате
	numUsers, err := buf.ReadByte()
	if err != nil {
		return chat, err
	}

	// Пользователи в чате
	for i := 0; i < int(numUsers); i++ {
		var user storage.UserModel

		// Чтение длины логина пользователя
		loginLen, err := buf.ReadByte()
		if err != nil {
			return chat, err
		}

		// Чтение логина пользователя
		loginBytes := make([]byte, loginLen)
		if _, err := buf.Read(loginBytes); err != nil {
			return chat, err
		}
		user.Login = string(loginBytes)

		// Добавляем пользователя к чату
		chat.Users = append(chat.Users, user)
	}

	return chat, nil
}

func deserializeTextMessage(data []byte) (storage.MessageModel, error) {
	var message storage.MessageModel

	buf := bytes.NewReader(data)

	// Пропускаем тип события
	if _, err := buf.ReadByte(); err != nil {
		return message, err
	}

	// Чтение Id чата
	if err := binary.Read(buf, binary.BigEndian, &message.ChatId); err != nil {
		return message, err
	}

	// Чтение длины сообщения
	var messageLength int16
	if err := binary.Read(buf, binary.BigEndian, &messageLength); err != nil {
		return message, err
	}

	// Чтение сообщения
	messageBytes := make([]byte, messageLength)
	if _, err := buf.Read(messageBytes); err != nil {
		return message, err
	}
	message.Message = string(messageBytes)

	return message, nil
}

func serializeFileMessage(data []byte, fileSize int32) (storage.MessageModel, int32, error) {
	var message storage.MessageModel

	buf := bytes.NewReader(data)

	// Пропускаем тип события
	if _, err := buf.ReadByte(); err != nil {
		return message, 0, err
	}

	// Чтение Id чата
	if err := binary.Read(buf, binary.BigEndian, &message.ChatId); err != nil {
		return message, 0, err
	}

	// Чтение Id сообщения
	if err := binary.Read(buf, binary.BigEndian, &message.Id); err != nil {
		return message, 0, err
	}

	// Чтение длины названия файла
	var fileNameLength uint16
	if err := binary.Read(buf, binary.BigEndian, &fileNameLength); err != nil {
		return message, 0, err
	}

	// Чтение названия файла
	fileNameBytes := make([]byte, fileNameLength)
	if _, err := buf.Read(fileNameBytes); err != nil {
		return message, 0, err
	}
	message.Message = string(fileNameBytes)

	return message, int32(fileSize), nil
}

func deserializeEventLeaveChat(data []byte) (int64, error) {
	// Создание буфера для десериализации
	buffer := bytes.NewReader(data)

	// Чтение типа события (пропускаем, т.к. он уже известен)
	if _, err := buffer.ReadByte(); err != nil {
		return 0, err
	}

	// Чтение ID чата
	var chatID int64
	if err := binary.Read(buffer, binary.BigEndian, &chatID); err != nil {
		return 0, err
	}

	return chatID, nil
}

func deserializeEventInviteChat(data []byte) (int64, string, error) {
	// Создание буфера для десериализации
	buffer := bytes.NewReader(data)

	// Чтение типа события (пропускаем, т.к. он уже известен)
	if _, err := buffer.ReadByte(); err != nil {
		return 0, "", err
	}

	// Чтение ID чата
	var chatID int64
	if err := binary.Read(buffer, binary.BigEndian, &chatID); err != nil {
		return 0, "", err
	}

	// Чтение длины логина пользователя
	userLoginLength, err := buffer.ReadByte()
	if err != nil {
		return 0, "", err
	}

	// Чтение логина пользователя
	userLoginBytes := make([]byte, userLoginLength)
	if _, err := buffer.Read(userLoginBytes); err != nil {
		return 0, "", err
	}
	userLogin := string(userLoginBytes)

	return chatID, userLogin, nil
}

func (m *Messager) serealizeEvendAddToChat(chat storage.ChatModel) ([]byte, error) {
	var buf bytes.Buffer

	// Номер команды
	buf.WriteByte(4)
	b, err := m.serializeChat(chat)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	return buf.Bytes(), nil
}
