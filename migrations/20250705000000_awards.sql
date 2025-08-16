-- +goose Up
-- +goose StatementBegin

-- Dictionary of award types
CREATE TABLE awards (
    id TEXT PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE, -- e.g., product_of_day, product_of_week, product_of_month
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(255), -- optional icon name or URL
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Link table between launches and awards
CREATE TABLE launch_awards (
    id TEXT PRIMARY KEY,
    launch_id TEXT NOT NULL REFERENCES launches(id) ON DELETE CASCADE,
    award_id TEXT NOT NULL REFERENCES awards(id) ON DELETE CASCADE,
    period_date DATE NOT NULL, -- canonical date representing the period (day date, week's Monday, month's 1st)
    awarded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(award_id, period_date) -- only one winner per award and period
);
CREATE INDEX idx_launch_awards_launch_id ON launch_awards(launch_id);
CREATE INDEX idx_launch_awards_award_period ON launch_awards(award_id, period_date);

-- seed common awards (simple string IDs)
INSERT INTO awards (id, code, name, description, icon) VALUES
    ('day', 'product_of_day', 'Продукт дня', 'Лучший запуск за день', '🏅'),
    ('week', 'product_of_week', 'Продукт недели', 'Лучший запуск за неделю', '🥇'),
    ('month', 'product_of_month', 'Продукт месяца', 'Лучший запуск за месяц', '🏆'),
    ('year', 'product_of_year', 'Продукт года', 'Лучший запуск за год', '🎖️');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS launch_awards;
DROP TABLE IF EXISTS awards;
-- +goose StatementEnd


