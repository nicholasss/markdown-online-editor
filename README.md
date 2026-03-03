# Markdown Online Editor

A simple and robust online Markdown editor.

## Requirements

- [Goose](https://github.com/pressly/goose): A database migration tool.  
Supports SQL migrations and Go functions.
- [Sqlite 3](https://www.sqlite.org/): A simple file-based database.

## Setup

### Directories

- `sql/schema`: Where the SQL migrations are stored.

### Steps

1. From the root directory create `./notes.db` and run `goose up`. You should  
see it successfully migrate to fully through, without failures.
2.
