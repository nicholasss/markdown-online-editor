-- +goose Up
ALTER TABLE notes
ADD COLUMN note_title TEXT;

-- +goose Down
ALTER TABLE notes
DROP COLUMN note_title;
