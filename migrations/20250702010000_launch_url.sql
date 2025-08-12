-- +goose Up
-- +goose StatementBegin
ALTER TABLE launches ADD COLUMN image_url TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite does not support DROP COLUMN; keeping the column.
-- +goose StatementEnd
