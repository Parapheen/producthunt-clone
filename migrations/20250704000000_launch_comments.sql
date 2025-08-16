-- +goose Up
-- +goose StatementBegin
CREATE TABLE launch_comments (
    id TEXT PRIMARY KEY,
    launch_id TEXT NOT NULL REFERENCES launches(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id TEXT NULL REFERENCES launch_comments(id) ON DELETE CASCADE,
    content_html TEXT NOT NULL,
    tag TEXT NULL CHECK (tag IN ('idea','question','like')),
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_launch_comments_launch_id ON launch_comments(launch_id);
CREATE INDEX idx_launch_comments_parent_id ON launch_comments(parent_id);

CREATE TABLE launch_comment_upvotes (
    comment_id TEXT NOT NULL REFERENCES launch_comments(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (comment_id, user_id)
);
CREATE INDEX idx_comment_upvotes_comment_id ON launch_comment_upvotes(comment_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS launch_comment_upvotes;
DROP TABLE IF EXISTS launch_comments;
-- +goose StatementEnd


