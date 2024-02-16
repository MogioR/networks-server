package messager

import (
	"context"
	"fmt"
	"messager-server/internal/storage"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

type Messager struct {
	users        map[int64]net.Conn
	secureTokens map[string]int64
	storage      *storage.Storage

	tokensMutex sync.RWMutex
	usersMutex  sync.RWMutex
}

func NewMessager(storage *storage.Storage) *Messager {
	return &Messager{
		users:        make(map[int64]net.Conn, 0),
		secureTokens: make(map[string]int64, 0),
		storage:      storage,
	}
}

func (m *Messager) GetTokenForUser(userId int64) (string, error) {
	m.tokensMutex.Lock()
	defer m.tokensMutex.Unlock()

	if _, ok := m.users[userId]; !ok {
		token := uuid.New().String()
		m.secureTokens[token] = userId
		return token, nil
	} else {
		m.usersMutex.Lock()
		delete(m.users, userId)
		m.usersMutex.Unlock()

		token := uuid.New().String()
		m.secureTokens[token] = userId
		return token, nil
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

		key := make([]byte, 1024)
		n, err := conn.Read(key)
		if err != nil {
			errCh <- err
			return
		}

		m.tokensMutex.RLock()
		defer m.tokensMutex.RUnlock()

		if userId, ok := m.secureTokens[string(key[:n])]; ok {
			resCh <- userId
			return
		} else {
			errCh <- fmt.Errorf("wrong pass")
			return
		}
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
			continue
		}
		m.processEvent(response[:n])
	}

	fmt.Printf("User %d disconnected\n", userId)
	m.usersMutex.Lock()
	delete(m.users, userId)
	m.usersMutex.Unlock()
}

func (m *Messager) sendChats(conn net.Conn, userId int64) error {
	chats, err := m.storage.GetChats(userId)
	if err != nil {
		return err
	}

	if len(chats) > 256 {
		chats = chats[:256]
	}

	message, err := m.serializeChats(chats)
	if err != nil {
		return err
	}

	m.usersMutex.RLock()
	_, err = m.users[userId].Write(message)
	m.usersMutex.RUnlock()

	return err
}

func (m *Messager) processEvent(eventMsg []byte) error {
	switch eventMsg[0] {
	case 7:
		msg, err := deseralizeChatCreate(eventMsg)
		if err != nil {
			return err
		}

		userIds, err := m.storage.CreateChat(msg.Name, msg.Users)
		if err != nil {
			return err
		}

		for i, user := range msg.Users {
			user.Id = userIds[i]
		}

		message, err := m.serealizeEvendAddToChat(msg)
		if err != nil {
			return err
		}

		m.usersMutex.RLock()
		for _, id := range userIds {
			if _, ok := m.users[id]; ok {
				_, err = m.users[id].Write(message)
				if err != nil {
					log.Warn(err)
				}
			}
		}
		m.usersMutex.RUnlock()

	default:
		return fmt.Errorf("unknown event type")
	}

	return nil
}
