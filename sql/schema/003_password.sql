-- +goose Up
ALTER TABLE users ADD hashed_password text NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLE users DROP COLUMN hashed_password;
