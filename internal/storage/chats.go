package storage

import (
	"database/sql"
	"fmt"
	"time"
)

func (s *Storage) GetChatUsers(chatID int64) ([]UserModel, error) {
	query := `
		SELECT u.id, u.LOGIN FROM UsersChats uc
		INNER JOIN Users u on uc.USER_ID = u.id
		WHERE uc.CHAT_ID = $1
	`

	rows, err := s.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userId int64
	var userName string

	result := make([]UserModel, 0)
	for rows.Next() {
		err = rows.Scan(&userId, &userName)
		if err != nil {
			return nil, err
		}
		result = append(result, UserModel{
			Id:    userId,
			Login: userName,
		})
	}

	return result, nil
}

func (s *Storage) GetChats(userId int64) ([]ChatModel, error) {
	query := `
		SELECT c.ID, c.NAME, MAX(cm.posted_at) as t FROM UsersChats uc
		INNER JOIN Users u on uc.USER_ID = u.id
		INNER JOIN Chats c on uc.CHAT_ID = c.id
		left JOIN ChatsMessages cm on uc.CHAT_ID = cm.CHAT_ID
		WHERE uc.USER_ID = $1
		group by c.ID, c.NAME
		order by t
	`

	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatId int64
	var chatName string
	var lastDateRaw *string
	var lastMessage time.Time

	result := make([]ChatModel, 0)
	for rows.Next() {
		err = rows.Scan(&chatId, &chatName, &lastDateRaw)
		if err != nil {
			return nil, err
		}

		if lastDateRaw != nil {
			lastMessage, err = time.Parse(time.RFC3339, *lastDateRaw)
			if err != nil {
				return nil, err
			}
		} else {
			lastMessage, _ = time.Parse("2006-01-02", "1001-01-01")
		}

		users, err := s.GetChatUsers(chatId)
		if err != nil {
			return nil, err
		}

		result = append(result, ChatModel{
			Id:          chatId,
			Name:        chatName,
			LastMessage: lastMessage,
			Users:       users,
		})
	}

	return result, nil
}

func (s *Storage) AddUserToChat(userId int64, chatId int64) error {
	query := `
		SELECT USER_ID
		FROM UsersChats 
		WHERE CHAT_ID = $1 and USER_ID = $2
	`
	var userIdBuf int64
	if err := s.db.QueryRow(query, chatId, userId).Scan(&userIdBuf); err != sql.ErrNoRows {
		if err != nil {
			return fmt.Errorf("add user %d to chat %d error -> %s", userId, chatId, err)
		}
	}

	if userIdBuf != 0 {
		return fmt.Errorf("user alrady in chat")
	}

	query = `
		INSERT INTO UsersChats (USER_ID, CHAT_ID) 
		VALUES($1, $2)
	`

	_, err := s.db.Exec(query, userId, chatId)

	return err
}

func (s *Storage) KickUserFromChat(userId int64, chatId int64) error {
	query := `
		DELETE UsersChats
		WHERE CHAT_ID = $2 and USER_ID = $1
	`

	_, err := s.db.Exec(query, userId, chatId)

	return err
}

func (s *Storage) CreateChat(name string, users []UserModel) (userIDs []int64, err error) {
	// Проверка наличия пользователей в списке
	if len(users) == 0 {
		return nil, fmt.Errorf("no users")
	}

	// Начало транзакции
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Вставка записи в таблицу Chats
	var chatID int64
	err = tx.QueryRow("INSERT INTO Chats(Name) VALUES($1) RETURNING ID", name).Scan(&chatID)
	if err != nil {
		return nil, err
	}

	// Проверка наличия пользователей и создание записей в таблице UsersChats
	for _, user := range users {
		var userID int64
		err := tx.QueryRow("SELECT ID FROM Users WHERE LOGIN = $1", user.Login).Scan(&userID)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unknown user '%s'", user.Login)
		} else if err != nil {
			return nil, err
		}

		_, err = tx.Exec("INSERT INTO UsersChats(USER_ID, CHAT_ID) VALUES($1, $2)", userID, chatID)
		if err != nil {
			return nil, err
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}
