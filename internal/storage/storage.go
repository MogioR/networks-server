package storage

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

func New(db *sql.DB, logger *logrus.Logger) *Storage {
	s := &Storage{
		db:     db,
		logger: logger,
	}
	return s
}

type Storage struct {
	db     *sql.DB
	logger *logrus.Logger
}
