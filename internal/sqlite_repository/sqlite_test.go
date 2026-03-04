package sqlite_test

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	note "github.com/nicholasss/markdown-online-editor/internal/models"
	sqlite "github.com/nicholasss/markdown-online-editor/internal/sqlite_repository"
)

const databaseString = ":memory:"

func checkNoteEquality(t *testing.T, got, want *note.Note) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("Note IDs do not match. got=%s want=%s", got.ID, want.ID)
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

func prepareTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Create the table to test against
	createQuery := `CREATE TABLE
notes (
	id TEXT PRIMARY KEY,
	created_at INTEGER,
	updated_at INTEGER,
	note_text BLOB,
	note_title TEXT
);`
	_, err := db.Exec(createQuery)
	if err != nil {
		t.Fatalf("Unable to create table: %s", err)
	}

	// Insert Query
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
	}

	for _, record := range insertData {
		_, err := db.Exec(insertQuery, record.id, record.createdAt, record.updatedAt, record.noteText, record.noteTitle)
		if err != nil {
			t.Fatalf("Error preparing database with test data: %s\n", err)
		}
	}
}

func TestInsertNote(t *testing.T) {
	testTable := []struct {
		name        string
		newNote     *note.Note
		expectError bool
		wantError   error
		wantNote    *note.Note
	}{
		{
			name: "valid-1",
			newNote: &note.Note{
				ID: uuid.MustParse("2037225a-da01-4609-ad78-fb37c3f6cf06"),
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
			expectError: false,
			wantError:   nil,
			wantNote: &note.Note{
				ID: uuid.MustParse("2037225a-da01-4609-ad78-fb37c3f6cf06"),
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
			// setup repository
			repo, err := sqlite.NewSqliteRepository(databaseString)
			if err != nil {
				t.Fatalf("Failed to setup in-memory sqlite3 repository: %s\n", err)
			}

			// run prepare test database
			prepareTestDB(t, repo.DB)

			// defer teardown
			defer func() {
				err := repo.DB.Close()
				if err != nil {
					t.Errorf("Unable to close in-memory sqlite3 repo: %s\n", err)
				}
			}()

			// call the test function
			gotNote, gotErr := repo.InsertNote(t.Context(), testCase.newNote)

			// checking if we expect an error
			if (gotErr != nil) != testCase.expectError {
				t.Errorf("InsertNote() got error: '%v', expected error: '%v'", gotErr, testCase.wantError)
			}

			// checking error type if its not nil
			if gotErr != nil {
				if !errors.Is(gotErr, testCase.wantError) {
					t.Errorf("got error: %v, want error: %v", gotErr, testCase.wantError)
				}
			}

			// check returned note
			if !testCase.expectError && gotNote != nil {
				checkNoteEquality(t, gotNote, testCase.wantNote)
			}
		})
	}
}
