-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN bio TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Note: SQLite doesn't support DROP COLUMN easily, so we'll keep the column
-- +goose StatementEnd
