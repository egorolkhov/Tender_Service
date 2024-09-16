-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создание типа для статуса предложений
CREATE TYPE bid_status AS ENUM ('Created', 'Published', 'Canceled', 'Approved', 'Rejected');

-- Создание типа для авторов предложений
CREATE TYPE author_type AS ENUM ('User', 'Organization');

-- Основная таблица для хранения актуальных данных предложения
CREATE TABLE IF NOT EXISTS bid (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    status bid_status NOT NULL,
    tender_id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
    author_type author_type NOT NULL,
    author_id UUID NOT NULL REFERENCES employee(id),
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица для хранения всех версий предложений
CREATE TABLE IF NOT EXISTS bid_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bid_id UUID NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    status bid_status NOT NULL,
    tender_id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
    author_type author_type NOT NULL,
    author_id UUID NOT NULL REFERENCES employee(id),
    version INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX idx_bid_history_bid_id ON bid_history (bid_id);
CREATE INDEX idx_bid_history_version ON bid_history (bid_id, version);

-- +goose Down
DROP INDEX IF EXISTS idx_bid_history_version;
DROP INDEX IF EXISTS idx_bid_history_bid_id;
DROP TABLE IF EXISTS bid_history;
DROP TABLE IF EXISTS bid;
DROP TYPE IF EXISTS bid_status;
DROP TYPE IF EXISTS author_type;
