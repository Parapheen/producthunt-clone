-- +goose Up
-- +goose StatementBegin

-- Users (Created first)
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Social Accounts (Created after users)
CREATE TABLE social_accounts (
    id TEXT PRIMARY KEY,
    provider VARCHAR(255) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, provider_id)
);
CREATE INDEX idx_social_accounts_user_id ON social_accounts(user_id);

-- Sessions (Created after users)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);


CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    tagline VARCHAR(255),
    url VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_members (
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (product_id, user_id)
);

CREATE TABLE launches (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tagline VARCHAR(255),
    state VARCHAR(255) NOT NULL CHECK(state IN ('draft', 'review', 'declined', 'published', 'archived')),
    url VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    launch_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (product_id, slug)
);
CREATE INDEX idx_launches_product_id ON launches(product_id);

CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Note: The INSERT statement now omits the ID, as it's auto-generated
INSERT INTO categories (id, name, slug) VALUES
(NULL, 'AI', 'ai'),
(NULL, 'Аналитика', 'analytics'),
(NULL, 'Искусство', 'art'),
(NULL, 'Книги', 'books'),
(NULL, 'Поддержка', 'support'),
(NULL, 'Дизайн', 'design'),
(NULL, 'Инструменты разработчика', 'dev-tools'),
(NULL, 'E-commerce', 'ecommerce'),
(NULL, 'Образование', 'education'),
(NULL, 'Мероприятия', 'events'),
(NULL, 'Мода', 'fashion'),
(NULL, 'Fintech', 'fintech'),
(NULL, 'Игры', 'games'),
(NULL, 'Спорт', 'sport'),
(NULL, 'Найм', 'recruiting'),
(NULL, 'Инвестиции', 'investing'),
(NULL, 'Маркетплейс', 'marketplace'),
(NULL, 'Маркетинг', 'marketing'),
(NULL, 'Мессенджеры', 'messengers'),
(NULL, 'Музыка', 'music'),
(NULL, 'Платежи', 'payments'),
(NULL, 'Продуктивность', 'productivity'),
(NULL, 'Продажи', 'sales'),
(NULL, 'Безопасность', 'security'),
(NULL, 'Индихакеры', 'indiehackers'),
(NULL, 'Legaltech', 'legaltech'),
(NULL, 'Здоровье', 'health'),
(NULL, 'Путешествия', 'travel');

CREATE TABLE product_categories (
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE, -- Correct data type and reference
    PRIMARY KEY (product_id, category_id)
);

CREATE TABLE launch_upvotes (
    launch_id TEXT NOT NULL REFERENCES launches(id) ON DELETE CASCADE, -- Corrected data type
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (launch_id, user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop tables in the reverse order of creation to respect dependencies.
DROP TABLE IF EXISTS launch_upvotes;
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS launches;
DROP TABLE IF EXISTS product_members;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS social_accounts;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
