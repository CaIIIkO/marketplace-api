-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_advertisements_price_kopecks ON advertisements(price_kopecks);
CREATE INDEX IF NOT EXISTS idx_advertisements_created ON advertisements(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_advertisements_price_kopecks;
DROP INDEX IF EXISTS idx_advertisements_created;
-- +goose StatementEnd