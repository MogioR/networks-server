package storage

import (
	"time"
)

func (s *Storage) GetMessagesFromChat(chatID int64) ([]MessageModel, error) {
	query := `
		SELECT cm.id, cm.USER_ID, cm.IS_FILE, cm.MESSAGE, cm.READED_AT, cm.posted_at FROM ChatsMessages cm
		WHERE cm.CHAT_ID = $1
		ORDER BY cm.posted_at
	`

	rows, err := s.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messageId, userId int64
	var isFile bool
	var message, readedAtRaw, postedAtRaw string

	result := make([]MessageModel, 0)
	for rows.Next() {
		err = rows.Scan(&messageId, &userId, &isFile, &message, &readedAtRaw, &postedAtRaw)
		if err != nil {
			return nil, err
		}

		readedAt, err := time.Parse(time.RFC3339, readedAtRaw)
		if err != nil {
			return nil, err
		}

		postedAt, err := time.Parse(time.RFC3339, postedAtRaw)
		if err != nil {
			return nil, err
		}

		result = append(result, MessageModel{
			Id:       messageId,
			UserId:   userId,
			IsFile:   isFile,
			Message:  message,
			ReadedAt: readedAt,
			PostedAt: postedAt,
		})
	}

	return result, nil
}

func (s *Storage) AddMessage(userId, chatId int64, message string, is_file bool) (int64, error) {
	query := `
		INSERT INTO ChatsMessages (USER_ID, CHAT_ID, MESSAGE, IS_FILE) 
		VALUES($1, $2, $3, $4) 
		RETURNING id
	`
	var messageId int64
	err := s.db.QueryRow(query, userId, chatId, message, is_file).Scan(&messageId)
	return messageId, err
}

func (s *Storage) MessageAreReaded(messageId, chatId int64) error {
	query := `
		Update ChatsMessages
		set READED_AT = NOW()
		where ID = $1 and CHAT_ID = $2
	`

	_, err := s.db.Exec(query, messageId, chatId)
	return err
}
