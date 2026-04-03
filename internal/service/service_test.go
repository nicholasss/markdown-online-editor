// Package service_test is to test the service package.
package service_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	note "github.com/nicholasss/markdown-online-editor/internal/models"
	"github.com/nicholasss/markdown-online-editor/internal/service"
	sqlite "github.com/nicholasss/markdown-online-editor/internal/sqlite_repository"
)

// === Testing Constants ===

const databaseString = ":memory:"

// === Testing Helper Functions ===

func checkNoteEquality(t *testing.T, skipID bool, got, want *note.Note) {
	t.Helper()

	if !skipID {
		if got.ID != want.ID {
			t.Errorf("Note IDs do not match. got=%s want=%s", got.ID, want.ID)
		}
	}
	if !bytes.Equal(got.NoteText, want.NoteText) {
		diff := cmp.Diff(string(want.NoteText), string(got.NoteText))
		t.Errorf("Note Texts do not match (-want +got):\n%s", diff)
	}
	if got.NoteTitle != want.NoteTitle {
		t.Errorf("Note Titles does not match. got=%s want=%s", got.NoteTitle, want.NoteTitle)
	}

	if !want.NoteCreatedAt.IsZero() {
		if !want.NoteCreatedAt.Equal(got.NoteCreatedAt) {
			t.Errorf("Note CreatedAts do not match. got=%d want=%d", got.NoteCreatedAt.Unix(), want.NoteCreatedAt.Unix())
		}
	}

	if !want.NoteUpdatedAt.IsZero() {
		if !want.NoteUpdatedAt.Equal(got.NoteUpdatedAt) {
			t.Errorf("Note UpdatedAts do not match. got=%d want=%d", got.NoteUpdatedAt.Unix(), want.NoteUpdatedAt.Unix())
		}
	}
}

// Helper function to check the error values for a particular case
func checkTestError(t *testing.T, shouldErr bool, gotErr, wantErr error) {
	t.Helper()

	// checking if we expect an error
	if (gotErr != nil) != shouldErr {
		t.Errorf("Test Failed: got error: '%v', expected error: '%v'", gotErr, wantErr)
	}

	// checking error type if its not nil
	if gotErr != nil {
		if !errors.Is(gotErr, wantErr) {
			t.Errorf("Test Failed: got error: %v, want error: %v", gotErr, wantErr)
		}
	}
}

// Defered function to close the database
func closeTestService(t *testing.T, serv *service.ServiceImplementation) {
	t.Helper()

	repo, isSqliteRepo := serv.Repo.(*sqlite.SqliteRepository)
	if !isSqliteRepo {
		t.Errorf("Unexpected repository implementation.")
	}

	err := repo.CloseRepository()
	if err != nil {
		t.Errorf("Unable to close in-memory sqlite3 repo: %s\n", err)
	}
}

func newTestService(t *testing.T) *service.ServiceImplementation {
	t.Helper()

	// Create repository in memory
	repo, err := sqlite.NewSqliteRepository(databaseString)
	if err != nil {
		t.Fatalf("Failed to setup in-memory sqlite3 repository: %s\n", err)
	}

	// Create the table to test against
	createQuery := `CREATE TABLE
notes (
	id TEXT PRIMARY KEY,
	created_at INTEGER,
	updated_at INTEGER,
	note_text BLOB,
	note_title TEXT
);`
	_, err = repo.DB.Exec(createQuery)
	if err != nil {
		t.Fatalf("Unable to create table: %s", err)
	}

	// Insert query and mock data for database
	insertQuery := `INSERT INTO
notes (
	id,
	created_at,
	updated_at,
	note_text,
	note_title
) values (
	?,
	?,
	?,
	?,
	?
);`

	insertData := []struct {
		id        uuid.UUID
		createdAt int64
		updatedAt int64
		noteText  []byte
		noteTitle string
	}{
		{
			id:        uuid.MustParse("8050cf47-3145-4758-ac73-5ed384f5bd16"),
			createdAt: 1772637646,
			updatedAt: 1772638846,
			noteText: []byte(`# Baking notes
## First look at the steps
There are usually **lots of complex steps** in a recipe, and you must be prepared.
## Watch the oven
Make sure to not leave your baked goods in the oven, *without* keeping an eye on them. Especially ones that burn easily.
## Mind the ingredient amounts
Ingredients are sometimes extremely important to keep in specific ratios. Sometimes salted butter needs to be accounted for.
## Be open to trying new cuisines
There are sometimes great recipes from other cuisines and cultures that you would have never tried otherwise.`),
			noteTitle: "Baking Notes",
		},
		{
			id:        uuid.MustParse("337b8543-1272-4616-b9a3-3a16e5f9a522"),
			createdAt: 1772551246,
			updatedAt: 1772556346,
			noteText: []byte(`# Coding notes
## Watch out for typos
Typos within code can lead to annoying bugs, make sure you are practicing for accuracy, not just speed.
## Markup languages are your friend
Markup languages can be very useful when keeping notes or storing information in a document.
## Keep practicing
The worst thing you can do is stop learning and stop practicing.

<span id="counter">4</span>`),
			noteTitle: "Coding Notes",
		},
		{
			id:        uuid.MustParse("da0c2260-1a6f-4f49-837b-40831225dda9"),
			createdAt: 1772530246,
			updatedAt: 1772534346,
			noteText: []byte(`# Todo
Snowstorm is coming in Sunday morning, so I can do some of the preparation then.
However I need to go shopping Saturday to prepare.

- [x] get milk
- [x] get eggs
- [x] get bread
- [ ] get cat litter
- [ ] get driveway salt
- [ ] salt driveway and sidewalk`),
			noteTitle: "Short Todo",
		},
	}

	// Actually insert the mock data
	for _, record := range insertData {
		_, err := repo.DB.Exec(insertQuery, record.id, record.createdAt, record.updatedAt, record.noteText, record.noteTitle)
		if err != nil {
			t.Fatalf("Error preparing database with test data: %s\n", err)
		}
	}

	// Return prepared service for testing
	serv, err := service.NewService(repo)
	if err != nil {
		t.Fatalf("unable to setup service for testing due to: %s", err)
	}
	return serv
}

// === Method Testing ===

func TestCreateNote(t *testing.T) {
	testTable := []struct {
		name        string
		newNote     *note.Note
		shouldErr   bool
		wantErr     error
		skipIDCheck bool
		wantNote    *note.Note
	}{
		{
			name: "successful creating new note",
			newNote: &note.Note{
				NoteText: []byte(`# Language Learning
## Language List
- German: Fluent-ish
- Spanish: New
- Mandarin: No experience
## Goals
Learn German to C1
Learn Spanish to B1
Learn Mandarin to basic A1
## Priorities
German is top priority, with Spanish close behind. Mandarin does not matter.`),
				NoteTitle: "Language Learning",
			},
			shouldErr:   false,
			wantErr:     nil,
			skipIDCheck: true,
			wantNote: &note.Note{
				NoteText: []byte(`# Language Learning
## Language List
- German: Fluent-ish
- Spanish: New
- Mandarin: No experience
## Goals
Learn German to C1
Learn Spanish to B1
Learn Mandarin to basic A1
## Priorities
German is top priority, with Spanish close behind. Mandarin does not matter.`),
				NoteTitle: "Language Learning",
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup test serv
			serv := newTestService(t)
			defer closeTestService(t, serv)

			// call the test function
			createdAt := testCase.newNote.NoteCreatedAt
			title := testCase.newNote.NoteTitle
			text := testCase.newNote.NoteText
			gotNote, gotErr := serv.CreateNote(t.Context(), createdAt, title, text)

			// Check the returned errors
			checkTestError(t, testCase.shouldErr, gotErr, testCase.wantErr)

			// check returned note
			if !testCase.shouldErr && gotNote != nil {
				checkNoteEquality(t, testCase.skipIDCheck, gotNote, testCase.wantNote)
			}
		})
	}
}

func TestGetNote(t *testing.T) {
	testTable := []struct {
		name        string
		inputID     uuid.UUID
		shouldErr   bool
		wantErr     error
		skipIDCheck bool
		wantNote    *note.Note
	}{
		{
			name:        "successful getting existing note",
			inputID:     uuid.MustParse("8050cf47-3145-4758-ac73-5ed384f5bd16"),
			shouldErr:   false,
			wantErr:     nil,
			skipIDCheck: false,
			wantNote: &note.Note{
				ID:            uuid.MustParse("8050cf47-3145-4758-ac73-5ed384f5bd16"),
				NoteCreatedAt: time.Unix(1772637646, 0),
				NoteUpdatedAt: time.Unix(1772638846, 0),
				NoteText: []byte(`# Baking notes
## First look at the steps
There are usually **lots of complex steps** in a recipe, and you must be prepared.
## Watch the oven
Make sure to not leave your baked goods in the oven, *without* keeping an eye on them. Especially ones that burn easily.
## Mind the ingredient amounts
Ingredients are sometimes extremely important to keep in specific ratios. Sometimes salted butter needs to be accounted for.
## Be open to trying new cuisines
There are sometimes great recipes from other cuisines and cultures that you would have never tried otherwise.`),
				NoteTitle: "Baking Notes",
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// run test
			serv := newTestService(t)
			defer closeTestService(t, serv)

			// call the test function
			gotNote, gotErr := serv.GetNote(t.Context(), testCase.inputID)

			// Check the returned errors
			checkTestError(t, testCase.shouldErr, gotErr, testCase.wantErr)

			// check returned note
			if !testCase.shouldErr && gotNote != nil {
				checkNoteEquality(t, testCase.skipIDCheck, gotNote, testCase.wantNote)
			}
		})
	}
}

func TestGetAllNotes(t *testing.T) {
	testTable := []struct {
		name        string
		shouldErr   bool
		wantErr     error
		skipIDCheck bool
		wantNotes   []note.Note
	}{
		{
			name:        "successful getting all notes",
			shouldErr:   false,
			wantErr:     nil,
			skipIDCheck: false,
			wantNotes: []note.Note{
				{
					ID:            uuid.MustParse("8050cf47-3145-4758-ac73-5ed384f5bd16"),
					NoteCreatedAt: time.Unix(1772637646, 0),
					NoteUpdatedAt: time.Unix(1772638846, 0),
					NoteText: []byte(`# Baking notes
## First look at the steps
There are usually **lots of complex steps** in a recipe, and you must be prepared.
## Watch the oven
Make sure to not leave your baked goods in the oven, *without* keeping an eye on them. Especially ones that burn easily.
## Mind the ingredient amounts
Ingredients are sometimes extremely important to keep in specific ratios. Sometimes salted butter needs to be accounted for.
## Be open to trying new cuisines
There are sometimes great recipes from other cuisines and cultures that you would have never tried otherwise.`),
					NoteTitle: "Baking Notes",
				},
				{
					ID:            uuid.MustParse("337b8543-1272-4616-b9a3-3a16e5f9a522"),
					NoteCreatedAt: time.Unix(1772551246, 0),
					NoteUpdatedAt: time.Unix(1772556346, 0),
					NoteText: []byte(`# Coding notes
## Watch out for typos
Typos within code can lead to annoying bugs, make sure you are practicing for accuracy, not just speed.
## Markup languages are your friend
Markup languages can be very useful when keeping notes or storing information in a document.
## Keep practicing
The worst thing you can do is stop learning and stop practicing.

<span id="counter">4</span>`),
					NoteTitle: "Coding Notes",
				},
				{
					ID:            uuid.MustParse("da0c2260-1a6f-4f49-837b-40831225dda9"),
					NoteCreatedAt: time.Unix(1772530246, 0),
					NoteUpdatedAt: time.Unix(1772534346, 0),
					NoteText: []byte(`# Todo
Snowstorm is coming in Sunday morning, so I can do some of the preparation then.
However I need to go shopping Saturday to prepare.

- [x] get milk
- [x] get eggs
- [x] get bread
- [ ] get cat litter
- [ ] get driveway salt
- [ ] salt driveway and sidewalk`),
					NoteTitle: "Short Todo",
				},
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// run test
			serv := newTestService(t)
			defer closeTestService(t, serv)

			// call the test function
			gotNotes, gotErr := serv.GetAllNotes(t.Context())

			// Check the returned errors
			checkTestError(t, testCase.shouldErr, gotErr, testCase.wantErr)

			// check returned note
			for i, gotNote := range *gotNotes {
				if !testCase.shouldErr {
					checkNoteEquality(t, testCase.skipIDCheck, &gotNote, &testCase.wantNotes[i])
				}
			}
		})
	}
}

func TestUpdateNote(t *testing.T) {
	testTable := []struct {
		name        string
		shouldErr   bool
		wantErr     error
		skipIDCheck bool
		updatedNote *note.Note
		wantNote    *note.Note
	}{
		{
			name:        "successful note text update",
			shouldErr:   false,
			wantErr:     nil,
			skipIDCheck: false,
			updatedNote: &note.Note{
				ID:        uuid.MustParse("8050cf47-3145-4758-ac73-5ed384f5bd16"),
				NoteText:  []byte(`new text`),
				NoteTitle: "Baking Notes",
			},
			wantNote: &note.Note{
				ID:        uuid.MustParse("8050cf47-3145-4758-ac73-5ed384f5bd16"),
				NoteText:  []byte(`new text`),
				NoteTitle: "Baking Notes",
			},
		},
		{
			name:        "successful note title update",
			shouldErr:   false,
			wantErr:     nil,
			skipIDCheck: false,
			updatedNote: &note.Note{
				ID: uuid.MustParse("da0c2260-1a6f-4f49-837b-40831225dda9"),
				NoteText: []byte(`# Todo
Snowstorm is coming in Sunday morning, so I can do some of the preparation then.
However I need to go shopping Saturday to prepare.

- [x] get milk
- [x] get eggs
- [x] get bread
- [ ] get cat litter
- [ ] get driveway salt
- [ ] salt driveway and sidewalk`),
				NoteTitle: "Actually medium length todo",
			},
			wantNote: &note.Note{
				ID: uuid.MustParse("da0c2260-1a6f-4f49-837b-40831225dda9"),
				NoteText: []byte(`# Todo
Snowstorm is coming in Sunday morning, so I can do some of the preparation then.
However I need to go shopping Saturday to prepare.

- [x] get milk
- [x] get eggs
- [x] get bread
- [ ] get cat litter
- [ ] get driveway salt
- [ ] salt driveway and sidewalk`),
				NoteTitle: "Actually medium length todo",
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// run test
			serv := newTestService(t)
			defer closeTestService(t, serv)

			// call the test function
			noteID := testCase.updatedNote.ID
			title := testCase.updatedNote.NoteTitle
			text := testCase.updatedNote.NoteText
			gotNote, gotErr := serv.UpdateNote(t.Context(), noteID, title, text)

			// Check the returned errors
			checkTestError(t, testCase.shouldErr, gotErr, testCase.wantErr)

			// check returned note
			if !testCase.shouldErr {
				checkNoteEquality(t, testCase.skipIDCheck, gotNote, testCase.wantNote)
			}
		})
	}
}

func TestDeleteNote(t *testing.T) {
	testTable := []struct {
		name         string
		shouldErr    bool
		wantErr      error
		deletingNote *note.Note
	}{
		{
			name:      "successful deletion of note",
			shouldErr: false,
			wantErr:   nil,
			deletingNote: &note.Note{
				ID: uuid.MustParse("da0c2260-1a6f-4f49-837b-40831225dda9"),
			},
		},
		{
			name:      "failing to delete nonexistent note",
			shouldErr: true,
			wantErr:   sqlite.ErrNoRowsDeleted,
			deletingNote: &note.Note{
				ID: uuid.New(),
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup testing repository
			serv := newTestService(t)
			defer closeTestService(t, serv)

			// perform test
			gotErr := serv.Repo.DeleteNote(t.Context(), testCase.deletingNote)

			// Check the returned errors
			checkTestError(t, testCase.shouldErr, gotErr, testCase.wantErr)
		})
	}
}
