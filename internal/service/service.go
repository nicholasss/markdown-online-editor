// Package service implements the service interface from the internal/notes.
//
// Since it is not relying on any particular tool, framework, or technology, it will be the only implementation.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	note "github.com/nicholasss/markdown-online-editor/internal/models"
)

// === Service Implementation ===

type ServiceImplementation struct {
	Repo note.NoteRepository
}

func NewService(repo note.NoteRepository) (*ServiceImplementation, error) {
	return &ServiceImplementation{Repo: repo}, nil
}

// Service Methods

// CreateNote will ensure that the note is valid and pass it onto the repository layer for insertion into the underlying database.
// createdAt will have already been populated within the handler when the create note endpoint is hit.
func (serv *ServiceImplementation) CreateNote(ctx context.Context, createdAt time.Time, title string, text []byte) (*note.Note, error) {
	if createdAt.After(time.Now().Add(time.Second)) || createdAt.IsZero() {
		return nil, errors.New("invalid creation time")
	}
	if title == "" {
		return nil, errors.New("empty title")
	}
	// max text size is handled in repository layer
	if len(text) <= 0 {
		return nil, errors.New("empty text")
	}

	noteToCreate := &note.Note{
		ID:            uuid.New(),
		NoteCreatedAt: createdAt,
		NoteUpdatedAt: createdAt,
		NoteText:      text,
		NoteTitle:     title,
	}
	createdNote, err := serv.Repo.InsertNote(ctx, noteToCreate)
	if err != nil {
		return nil, err
	}

	return createdNote, nil
}

func (serv *ServiceImplementation) GetNote(ctx context.Context, noteID uuid.UUID) (*note.Note, error) {
	return serv.Repo.GetNote(ctx, noteID)
}

func (serv *ServiceImplementation) GetAllNotes(ctx context.Context) (*[]note.Note, error) {
	return serv.Repo.GetAllNotes(ctx)
}

func (serv *ServiceImplementation) UpdateNote(ctx context.Context, noteID uuid.UUID, updatedAt time.Time, newTitle string, newText []byte) (*note.Note, error) {
	return nil, nil
}

func (serv *ServiceImplementation) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	// NOTE: for now... I would want to update the repo layer to only take UUID, instead of *note.Note
	validNoteToDelete, err := serv.Repo.GetNote(ctx, noteID)
	if err != nil {
		return err
	}

	return serv.Repo.DeleteNote(ctx, validNoteToDelete)
}
