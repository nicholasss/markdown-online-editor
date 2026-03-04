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

// === Sqlite Repository ===

func scanRow(row *sql.Row) (*note.Note, error) {
	// Declare query variables
	var queryID uuid.NullUUID
	var queryCreatedAt int64
	var queryUpdatedAt int64
	var queryNoteText []byte
	var queryNoteTitle string

	// Execute query with scan
	err := row.Scan(&queryID, &queryCreatedAt, &queryUpdatedAt, &queryNoteText, &queryNoteTitle)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to query row %w", err)
	} else if err != nil {
		log.Printf("Error scanning db row: %s\n", err)
		return nil, err
	}

	// Double check the queried ID
	if !queryID.Valid {
		return nil, errors.New("database returned a null UUID")
	}

	// Construct returning object
	queryNote := note.Note{
		ID:            queryID.UUID,
		NoteCreatedAt: time.Unix(queryCreatedAt, 0),
		NoteUpdatedAt: time.Unix(queryUpdatedAt, 0),
		NoteText:      queryNoteText,
		NoteTitle:     queryNoteTitle,
	}
	return &queryNote, nil
}

// === Sqlite Repository ===

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
	query := `INSERT INTO notes (
  	id,
  	created_at,
  	updated_at,
  	note_text,
		note_title
	) values (
		?,
		unixepoch(),
		unixepoch(),
		?,
		?
	) RETURNING *;`

	// Construct the row query and execute
	row := r.DB.QueryRowContext(ctx, query, newNote.ID, newNote.NoteText, newNote.NoteTitle)
	queryNote, err := scanRow(row)
	if err != nil {
		return nil, err
	}

	return queryNote, nil
}

func (r *SqliteRepository) GetNote(ctx context.Context, noteID uuid.UUID) (*note.Note, error) {
	if noteID == uuid.Nil {
		return nil, errors.New("nil uuid was passed in")
	}

	// Query literal
	query := `SELECT
  	id,
  	created_at,
  	updated_at,
  	note_text,
		note_title
	FROM
		notes
	WHERE
		id = ?
	RETURNING *;`

	// Construct the row query
	row := r.DB.QueryRowContext(ctx, query, noteID)
	queryNote, err := scanRow(row)
	if err != nil {
		return nil, err
	}

	return queryNote, nil
}

func (r *SqliteRepository) GetAllNotes(ctx context.Context) (*[]note.Note, error) {
	return nil, nil
}

func (r *SqliteRepository) UpdateNote(ctx context.Context, alteredNote *note.Note) (*note.Note, error) {
	if alteredNote == nil {
		return nil, errors.New("nil pointer provided")
	}

	// Query literal
	query := `UPDATE
		notes
	SET
  	updated_at = unixepoch(),
  	note_text = ?,
		note_title = ?
	WHERE
		id = ?
	RETURNING *;`

	// Construct the row query
	row := r.DB.QueryRowContext(ctx, query, alteredNote.NoteText, alteredNote.NoteTitle, alteredNote.ID)
	queryNote, err := scanRow(row)
	if err != nil {
		return nil, err
	}

	return queryNote, nil
}

func (r *SqliteRepository) DeleteNote(ctx context.Context, noteToDelete *note.Note) error {
	return nil
}
