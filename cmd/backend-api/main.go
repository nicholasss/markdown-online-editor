package main

import (
	"errors"
	"io/fs"
	"log"
	"os"

	"github.com/joho/godotenv"
	sqlite "github.com/nicholasss/markdown-online-editor/internal/sqlite_repository"
)

// Initialize and only call core functions
func main() {
	log.Println("Initializing Markdown Editor Backend...")

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s\n", err)
	}
	log.Println("Successfully loaded '.env' variables.")

	SqliteDBPath := os.Getenv("GOOSE_DBSTRING")

	_, err = os.Stat(SqliteDBPath)
	if errors.Is(err, fs.ErrNotExist) {
		log.Println("Database for notes was not found in root of project.")
		log.Println("You may need to create and initialize the database.")
		os.Exit(1)
	}
	if err != nil {
		log.Printf("Error checking status of notes database: %s\n", err)
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
