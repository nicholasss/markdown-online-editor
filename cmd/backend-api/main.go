package main

import (
	"errors"
	"io/fs"
	"log"
	"os"

	sqlite "github.com/nicholasss/markdown-online-editor/internal/sqlite_repository"
)

const SqliteDBPath = "./notes.db"

// Initialize and only call core functions
func main() {
	log.Println("Initializing Markdown Editor Backend...")

	_, err := os.Stat(SqliteDBPath)
	if errors.Is(err, fs.ErrNotExist) {
		log.Println("Database for notes was not found in root of project.")
		log.Println("You may need to create and initialize the database.")
		os.Exit(1)
	}
	if err != nil {
		log.Printf("Unknown error: %s\n", err)
		os.Exit(1)
	}

	repo, err := sqlite.NewSqliteRepository(SqliteDBPath)
	if err != nil {
		log.Println("Unable to setup Sqlite3 repository.")
		os.Exit(1)
	}

	defer func() {
		if err := repo.CloseRepository(); err != nil {
			log.Printf("Unable to close database: %s\n", err)
		}
		log.Println("Closed database successfully.")
	}()
}
