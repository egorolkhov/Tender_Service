-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создание типа для организации
CREATE TYPE organization_type AS ENUM ('IE', 'LLC', 'JSC');

-- Создание таблицы organization
CREATE TABLE IF NOT EXISTS organization (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type organization_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

-- Создание таблицы organization_responsible
CREATE TABLE IF NOT EXISTS organization_responsible (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE
    );

-- +goose Down
-- Откат миграции: удаление таблиц и типа
DROP TABLE IF EXISTS organization_responsible;
DROP TABLE IF EXISTS organization;
DROP TYPE IF EXISTS organization_type;
