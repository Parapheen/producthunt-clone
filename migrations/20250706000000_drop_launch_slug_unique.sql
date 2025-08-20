-- +goose Up
-- +goose StatementBegin

-- SQLite cannot drop a UNIQUE constraint directly; rebuild the table
PRAGMA foreign_keys=off;

BEGIN TRANSACTION;

-- Create new table without UNIQUE(product_id, slug)
CREATE TABLE launches_new (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tagline VARCHAR(255),
    state VARCHAR(255) NOT NULL CHECK(state IN ('draft', 'review', 'declined', 'published', 'archived')),
    url VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    image_url TEXT,
    launch_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy data
INSERT INTO launches_new (id, product_id, name, url, description, tagline, image_url, state, slug, launch_date, created_at, updated_at)
SELECT id, product_id, name, url, description, tagline, image_url, state, slug, launch_date, created_at, updated_at FROM launches;

-- Drop old table and rename new one
DROP TABLE launches;
ALTER TABLE launches_new RENAME TO launches;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_launches_product_id ON launches(product_id);

COMMIT;

PRAGMA foreign_keys=on;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Down migration is lossy; we reintroduce the UNIQUE(product_id, slug) constraint by rebuilding again
PRAGMA foreign_keys=off;
BEGIN TRANSACTION;

CREATE TABLE launches_old (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tagline VARCHAR(255),
    state VARCHAR(255) NOT NULL CHECK(state IN ('draft', 'review', 'declined', 'published', 'archived')),
    url VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    image_url TEXT,
    launch_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (product_id, slug)
);

INSERT INTO launches_old (id, product_id, name, url, description, tagline, image_url, state, slug, launch_date, created_at, updated_at)
SELECT id, product_id, name, url, description, tagline, image_url, state, slug, launch_date, created_at, updated_at FROM launches;

DROP TABLE launches;
ALTER TABLE launches_old RENAME TO launches;

CREATE INDEX IF NOT EXISTS idx_launches_product_id ON launches(product_id);

COMMIT;
PRAGMA foreign_keys=on;
-- +goose StatementEnd
