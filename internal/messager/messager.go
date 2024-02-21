package messager

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"messager-server/internal/messager/events"
	"messager-server/internal/storage"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Messager struct {
	users      map[int64]net.Conn
	storage    *storage.Storage
	usersMutex sync.RWMutex
}

func NewMessager(storage *storage.Storage) *Messager {
	return &Messager{
		users:   make(map[int64]net.Conn, 0),
		storage: storage,
	}
}

func (m *Messager) Auth(conn net.Conn) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	resCh := make(chan int64, 1)
	go func() {
		defer close(errCh)
		defer close(resCh)

		msgRaw := make([]byte, 1024)
		n, err := conn.Read(msgRaw)
		if err != nil {
			errCh <- err
			return
		}

		eventMsg := bytes.NewBuffer(msgRaw[:n])

		var authType int8
		err = binary.Read(eventMsg, binary.BigEndian, &authType)
		if err != nil || authType < 1 || authType > 2 {
			errCh <- fmt.Errorf("wrong event")
		}

		var userId int64
		if authType == 1 {
			event := events.RegisterEvent{}
			err = event.Deserialize(eventMsg)
			if err != nil {
				errCh <- fmt.Errorf("broken event")
			}
			userId, err = m.storage.Register(event.Login, event.Password)
		} else {
			event := events.LoginEvent{}
			err = event.Deserialize(eventMsg)
			if err != nil {
				errCh <- fmt.Errorf("broken event")
			}
			userId, err = m.storage.Auth(event.Login, event.Password)
		}
		if err != nil {
			errCh <- err
		}

		resCh <- userId
	}()

	select {
	case err := <-errCh:
		return -1, err
	case userId := <-resCh:
		return userId, nil
	case <-ctx.Done():
		return -1, fmt.Errorf("timeout auth")
	}
}

func (m *Messager) ConsumerHeandler(ctx context.Context, conn net.Conn, userId int64) {
	m.usersMutex.Lock()
	m.users[userId] = conn
	m.usersMutex.Unlock()

	m.usersMutex.RLock()
	err := m.sendChats(conn, userId)
	if err != nil {
		log.Warn(err)
	}
	m.usersMutex.RUnlock()

for_lable:
	for {
		select {
		case <-ctx.Done():
			break for_lable
		default:
		}

		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			break
		}
		buf := bytes.NewBuffer(response[:n])

		err = m.processEvent(buf, userId)
		if err != nil {
			event := events.SystemMessageEvent{
				Code:    1,
				Message: err.Error(),
			}
			conn.Write(event.Serialize().Bytes())
		}
	}

	fmt.Printf("User %d disconnected\n", userId)
	m.usersMutex.Lock()
	delete(m.users, userId)
	m.usersMutex.Unlock()
	conn.Close()
}

func (m *Messager) sendChats(conn net.Conn, userId int64) error {
	chats, err := m.storage.GetChats(userId)
	if err != nil {
		return err
	}

	if len(chats) > 256 {
		chats = chats[:256]
	}

	event := events.ChatEvent{}
	event.Chats = make([]*events.Chat, len(chats))
	for i, chat := range chats {
		event.Chats[i] = &events.Chat{
			ChatId:   chat.Id,
			ChatName: chat.Name,
		}

		event.Chats[i].Users = make([]events.User, len(chat.Users))
		for j, user := range chat.Users {
			event.Chats[i].Users[j] = events.User{
				Id:    user.Id,
				Login: user.Login,
			}
		}

		messages, err := m.storage.GetMessagesFromChat(chat.Id)
		if err != nil {
			return err
		}

		event.Chats[i].Messages = make([]events.Message, len(messages))
		for j, message := range messages {
			var isFileByte byte
			if message.IsFile {
				isFileByte = 1
			}

			event.Chats[i].Messages[j] = events.Message{
				Id:          message.Id,
				SenderId:    message.UserId,
				MessageType: isFileByte,
				Message:     message.Message,
				SendTime:    message.PostedAt,
				ReadTime:    message.ReadedAt,
			}
		}
	}

	message, err := event.Serialize()
	if err != nil {
		return err
	}

	m.usersMutex.RLock()
	_, err = m.users[userId].Write(message)
	m.usersMutex.RUnlock()

	return err
}

func (m *Messager) processEvent(eventMsg *bytes.Buffer, userId int64) error {
	var command int8
	err := binary.Read(eventMsg, binary.BigEndian, &command)
	if err != nil {
		return err
	}

	switch command {
	case 8:
		event := events.SendTextEvent{}
		err := event.Deserialize(eventMsg)
		if err != nil {
			return err
		}

		messageId, err := m.storage.AddMessage(userId, event.ChatId, event.Message, false)
		if err != nil {
			return err
		}

		users, err := m.storage.GetChatUsers(event.ChatId)
		if err != nil {
			return err
		}

		nilTime, _ := time.Parse("2006-01-02", "1001-01-01")
		answerEvent := events.NewMessageEvent{
			MessageId: messageId,
			ChatId:    event.ChatId,
			UserId:    userId,
			IsFile:    false,
			Message:   event.Message,
			SendTime:  time.Now(),
			ReadTime:  nilTime,
		}

		message, err := answerEvent.Serialize()
		if err != nil {
			return err
		}

		m.usersMutex.RLock()
		for _, user := range users {
			if _, ok := m.users[user.Id]; ok {
				_, err = m.users[user.Id].Write(message.Bytes())
				if err != nil {
					log.Warn(err)
				}
			}
		}
		m.usersMutex.RUnlock()
	case 9:
		event := events.SendFileInitEvent{}
		err := event.Deserialize(eventMsg)
		if err != nil {
			return err
		}

		file := make([]byte, event.FileSize)

		for i := 0; i < len(file); i += 1000 {
			top := i + 1000
			if top > len(file) {
				top = len(file)
			}
			m.usersMutex.RLock()
			_, err = m.users[userId].Read(file[i:top])
			m.usersMutex.RUnlock()
			if err != nil {
				return err
			}
		}

		messageId, err := m.storage.AddMessage(userId, event.ChatId, event.FileName, true)
		if err != nil {
			return err
		}

		err = SaveBytesToFile(fmt.Sprintf("files/%d_%d_%s", event.ChatId, messageId, event.FileName), file)
		if err != nil {
			return err
		}

		users, err := m.storage.GetChatUsers(event.ChatId)
		if err != nil {
			return err
		}

		nilTime, _ := time.Parse("2006-01-02", "1001-01-01")
		answerEvent := events.NewMessageEvent{
			MessageId: messageId,
			ChatId:    event.ChatId,
			UserId:    userId,
			IsFile:    true,
			Message:   event.FileName,
			SendTime:  time.Now(),
			ReadTime:  nilTime,
		}

		message, err := answerEvent.Serialize()
		if err != nil {
			return err
		}

		m.usersMutex.RLock()
		for _, user := range users {
			if _, ok := m.users[user.Id]; ok {
				_, err = m.users[user.Id].Write(message.Bytes())
				if err != nil {
					log.Warn(err)
				}
			}
		}
		m.usersMutex.RUnlock()
	case 13:
		event := events.GetFileFromChatEvent{}
		err := event.Deserialize(eventMsg)
		if err != nil {
			return err
		}

		messages, err := m.storage.GetMessagesFromChat(event.ChatId)
		if err != nil {
			return err
		}

		var fileName string
		for _, message := range messages {
			if message.Id == event.MessageId {
				fileName = message.Message
			}
		}

		file, err := ReadFileToBytes(fmt.Sprintf("files/%d_%d_%s", event.ChatId, event.MessageId, fileName))
		if err != nil {
			return err
		}

		answerEvent := events.SendFileToChatEvent{
			FileName: fileName,
			FileSize: int32(len(file)),
		}

		m.usersMutex.RLock()
		_, err = m.users[userId].Write(answerEvent.Serialize().Bytes())
		m.usersMutex.RUnlock()
		if err != nil {
			return err
		}
		for i := 0; i < len(file); i += 1000 {
			top := i + 1000
			if top > len(file) {
				top = len(file)
			}
			m.usersMutex.RLock()
			_, err = m.users[userId].Write(file[i:top])
			m.usersMutex.RUnlock()
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown event type")
	}

	return nil
}

func SaveBytesToFile(filename string, data []byte) error {
	// Используем функцию ioutil.WriteFile для записи данных в файл
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ReadFileToBytes(filePath string) ([]byte, error) {
	// Читаем содержимое файла в байтовый массив
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}
