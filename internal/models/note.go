package note

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Note represents a single note.
type Note struct {
	ID            uuid.UUID
	NoteCreatedAt time.Time
	NoteUpdatedAt time.Time
	NoteText      []byte
	NoteTitle     string
}

// NoteService is implemented by the note service within internal/services.
type NoteService interface {
	// Creates and returns a new note object after storing it in the repository.
	// This method is where a UUID is generated for the note.
	CreateNote(ctx context.Context, createdAt time.Time, title string, text []byte) (*Note, error)

	// Retrieves a stored note from the repository.
	GetNote(ctx context.Context, noteID uuid.UUID) (*Note, error)

	// Retrieves all notes stored in the repository.
	GetAllNotes(ctx context.Context) (*[]Note, error)

	// Updates specified note in the repository and returns the updated note.
	UpdateNote(ctx context.Context, noteID uuid.UUID, updatedAt time.Time, newTitle string, newText []byte) (*Note, error)

	// Deletes the specified note from within the repository.
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
}

// NoteRepository is implemented by the note repository within internal/repository with a concrete database.
type NoteRepository interface {
	// Creates a new note in the repository.
	InsertNote(ctx context.Context, newNote *Note) (*Note, error)

	// Retrieves note from the repository.
	GetNote(ctx context.Context, noteID uuid.UUID) (*Note, error)

	// Retrieves all notes from the repository.
	GetAllNotes(ctx context.Context) (*[]Note, error)

	// Updates the provided note, using its ID, within the repository.
	UpdateNote(ctx context.Context, alteredNote *Note) (*Note, error)

	// Deletes the specified note from within the repository.
	DeleteNote(ctx context.Context, noteToDelete *Note) error
}
