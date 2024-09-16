-- +goose Up
CREATE TABLE IF NOT EXISTS bid_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bid_id UUID NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
    review TEXT NOT NULL,
    reviewer UUID NOT NULL REFERENCES employee(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    bid_author_id UUID NOT NULL REFERENCES employee(id)
    );
-- +goose Down
DROP TABLE IF EXISTS bid_reviews;
