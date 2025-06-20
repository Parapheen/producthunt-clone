-- +goose Up
-- +goose StatementBegin

-- Users (Created first)
CREATE TABLE users (
    id TEXT PRIMARY KEY, -- Changed UUID to TEXT
    email VARCHAR(255) NOT NULL unique,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Changed NOW()
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP  -- Changed NOW()
);

-- Social Accounts (Created after users)
CREATE TABLE social_accounts (
    id TEXT PRIMARY KEY, -- Changed UUID to TEXT
    provider VARCHAR(255) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id), -- Changed UUID to TEXT
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Changed NOW()
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- Changed NOW()

    UNIQUE (provider, provider_id)
);

-- Sessions (Created after users)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY, -- Changed UUID to TEXT
    token VARCHAR(255) NOT NULL unique,
    user_id TEXT NOT NULL REFERENCES users(id), -- Changed UUID to TEXT
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Changed NOW()
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP  -- Changed NOW()
);

CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    url VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_members (
    product_id TEXT NOT NULL REFERENCES products(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    role VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (product_id, user_id)
);

CREATE TABLE launches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id TEXT NOT NULL REFERENCES products(id),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    tagline VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL, -- draft, in_review, declined, published, archived
    url VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    launch_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (product_id, slug)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop tables in reverse order of creation
DROP TABLE sessions;
DROP TABLE social_accounts; -- Corrected table name
DROP TABLE users;
DROP TABLE products;
DROP TABLE launches;
-- +goose StatementEnd
