package sqlite

import (
	"database/sql"

	"github.com/mahcks/serra/internal/db/repository"
	_ "github.com/mattn/go-sqlite3"
)

type Service interface {
	DB() *sql.DB
	Query() *repository.Queries
	Close() error
}

type sqliteService struct {
	db      *sql.DB
	queries *repository.Queries
}

func NewService(filepath string) (Service, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	queries := repository.New(db)

	return &sqliteService{
		db:      db,
		queries: queries,
	}, nil
}

func (s *sqliteService) DB() *sql.DB {
	return s.db
}

func (s *sqliteService) Query() *repository.Queries {
	return s.queries
}

func (s *sqliteService) Close() error {
	return s.db.Close()
}
