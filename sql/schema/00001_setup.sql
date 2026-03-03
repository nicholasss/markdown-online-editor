-- +goose Up
CREATE TABLE notes (
  id TEXT PRIMARY KEY,
  created_at INTEGER,
  updated_at INTEGER,
  note_text BLOB
);

-- +goose Down
DROP TABLE notes;
