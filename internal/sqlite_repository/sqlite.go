package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const DBDriver = "sqlite3"

func NewSqliteRepository(DBConnectionString string) (*SqliteRepository, error) {
	db, err := sql.Open(DBDriver, DBConnectionString)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, err
	}

	return &SqliteRepository{DB: db}, nil
}

type SqliteRepository struct {
	DB *sql.DB
}

func (repo *SqliteRepository) CloseRepository() error {
	return repo.DB.Close()
}
