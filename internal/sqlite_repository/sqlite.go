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

const (
	DBDriver = "sqlite3"
	MiB      = 1024 * 1024
)

var (
	// Nil Pointer Error
	ErrNilPointer = errors.New("nil pointer was provided")

	// ID Errors
	ErrInvalidIDProvided = errors.New("invalid note id was provided to repository")
	ErrInvalidIDReturned = errors.New("invalid note id was returned from database")

	// Note Data Errors
	ErrInvalidNoteText  = errors.New("invalid note text")
	ErrInvalidNoteTitle = errors.New("invalid note title")
	ErrInvalidCreatedAt = errors.New("invalid created at value ")
	ErrInvalidUpdatedAt = errors.New("invalid updated at value")
)

// === Sqlite Repository ===

// Used by InsertNote, UpdateNote, and GetNote since they all return the entire record
func scanRow(row *sql.Row) (*note.Note, error) {
	var queryID uuid.NullUUID
	var queryCreatedAt int64
	var queryUpdatedAt int64
	var queryNoteText []byte
	var queryNoteTitle string

	// Execute query with scan and checking return error
	err := row.Scan(&queryID, &queryCreatedAt, &queryUpdatedAt, &queryNoteText, &queryNoteTitle)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("unable to query row %w", err)
	} else if err != nil {
		log.Printf("Error scanning db row: %s\n", err)
		return nil, err
	}

	// Double check the queried ID
	if !queryID.Valid {
		return nil, ErrInvalidIDReturned
	}

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

func NewSqliteRepository(databaseString string) (*SqliteRepository, error) {
	db, err := sql.Open(DBDriver, databaseString)
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
	if newNote == nil {
		return nil, ErrNilPointer
	}
	if newNote.ID == uuid.Nil || newNote.ID == uuid.Max {
		return nil, ErrInvalidIDProvided
	}
	// Either no values, or is greater than 10 MiB in size
	if len(newNote.NoteText) == 0 || len(newNote.NoteText) > 10*MiB {
		return nil, ErrInvalidNoteText
	}
	if newNote.NoteTitle == "" {
		return nil, ErrInvalidNoteTitle
	}
	if !newNote.NoteCreatedAt.IsZero() {
		return nil, ErrInvalidCreatedAt
	}
	if !newNote.NoteUpdatedAt.IsZero() {
		return nil, ErrInvalidUpdatedAt
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

	row := r.DB.QueryRowContext(ctx, query, newNote.ID, newNote.NoteText, newNote.NoteTitle)
	queryNote, err := scanRow(row)
	if err != nil {
		return nil, err
	}

	return queryNote, nil
}

func (r *SqliteRepository) GetNote(ctx context.Context, noteID uuid.UUID) (*note.Note, error) {
	if noteID == uuid.Nil || noteID == uuid.Max {
		return nil, ErrInvalidIDProvided
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
		id = ?;`

	// Construct the row query
	row := r.DB.QueryRowContext(ctx, query, noteID)
	queryNote, err := scanRow(row)
	if err != nil {
		return nil, err
	}

	return queryNote, nil
}

func (r *SqliteRepository) GetAllNotes(ctx context.Context) (*[]note.Note, error) {
	query := `SELECT
  	id,
  	created_at,
  	updated_at,
  	note_text,
		note_title
	FROM
		notes;`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	// Scan the rows into a notes slice
	notes := make([]note.Note, 0)
	for rows.Next() {
		var queryID uuid.NullUUID
		var queryCreatedAt int64
		var queryUpdatedAt int64
		var queryNoteText []byte
		var queryNoteTitle string

		err := rows.Scan(&queryID, &queryCreatedAt, &queryUpdatedAt, &queryNoteText, &queryNoteTitle)
		if err != nil {
			return nil, err
		}

		// NOTE: Should I validate the queryID value returned from the database?
		// if !queryID.Valid { Do what if its invalid ?}

		queryNote := note.Note{
			ID:            queryID.UUID,
			NoteCreatedAt: time.Unix(queryCreatedAt, 0),
			NoteUpdatedAt: time.Unix(queryUpdatedAt, 0),
			NoteText:      queryNoteText,
			NoteTitle:     queryNoteTitle,
		}

		notes = append(notes, queryNote)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &notes, nil
}

func (r *SqliteRepository) UpdateNote(ctx context.Context, alteredNote *note.Note) (*note.Note, error) {
	if alteredNote == nil {
		return nil, ErrNilPointer
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
	if noteToDelete == nil {
		return ErrNilPointer
	}

	// Query literal
	query := `DELETE FROM
		notes
	WHERE
		id = ?;`

	// Run query
	res, err := r.DB.ExecContext(ctx, query, noteToDelete.ID)
	if err != nil {
		return err
	}

	// Check for no rows deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return errors.New("no rows deleted")
	}

	return nil
}
