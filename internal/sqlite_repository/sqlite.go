package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	note "github.com/nicholasss/markdown-online-editor/internal/models"
)

const DBDriver = "sqlite3"

func NewSqliteRepository(DBConnectionString string) (*SqliteRepository, error) {
	db, err := sql.Open(DBDriver, DBConnectionString)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return nil, err
	}

	return &SqliteRepository{DB: db}, nil
}

type SqliteRepository struct {
	DB *sql.DB
}

func (r *SqliteRepository) CloseRepository() error {
	return r.DB.Close()
}

func (r *SqliteRepository) InsertNote(ctx context.Context, newNote *note.Note) (*note.Note, error) {
	// Check for a nil pointer being passed in
	if newNote == nil {
		return nil, errors.New("nil pointer provided")
	}

	// Query literal
	query := `INSERT INTO expenses (
  	id,
  	created_at,
  	updated_at,
  	note_text
	) values (
		?,
		unixepoch(),
		unixepoch(),
		?
	) RETURNING *;`

	// Construct the row query
	row := r.DB.QueryRowContext(ctx, query, newNote.ID, newNote.NoteText)

	// Execute query with scan
	var queryID uuid.UUID
	var queryCreatedAt int64
	var queryUpdatedAt int64
	var queryNoteText []byte
	err := row.Scan(&queryID, &queryCreatedAt, &queryUpdatedAt, &queryNoteText)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to insert row %w", err)
	} else if err != nil {
		return nil, err
	}

	// Construct returning object
	queryNote := note.Note{
		ID:            queryID,
		NoteCreatedAt: time.Unix(queryCreatedAt, 0),
		NoteUpdatedAt: time.Unix(queryUpdatedAt, 0),
		NoteText:      queryNoteText,
	}
	return &queryNote, nil
}

func (r *SqliteRepository) QueryNote(ctx context.Context, noteID uuid.UUID) (*note.Note, error) {
	return nil, nil
}

func (r *SqliteRepository) QueryAllNotes(ctx context.Context) (*[]note.Note, error) {
	return nil, nil
}

func (r *SqliteRepository) AlterNote(ctx context.Context, note *note.Note) (*note.Note, error) {
	return nil, nil
}

func (r *SqliteRepository) DeleteNote(ctx context.Context, note *note.Note) error {
	return nil
}
