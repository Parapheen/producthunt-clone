-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN avatar_url TEXT;
ALTER TABLE products ADD COLUMN image_url TEXT;

CREATE TABLE launch_media (
    id TEXT PRIMARY KEY,
    launch_id TEXT NOT NULL REFERENCES launches(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_launch_media_launch_id ON launch_media(launch_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE launch_media;
-- Note: keep columns in place as other code may rely on them; SQLite doesn't support DROP COLUMN easily.
-- +goose StatementEnd


