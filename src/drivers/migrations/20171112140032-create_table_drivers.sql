
-- +migrate Up
CREATE TABLE drivers (
    id             bigserial PRIMARY KEY,
    name           text NOT NULL,
    license_number text NOT NULL UNIQUE
);

-- +migrate Down
DROP TABLE drivers;
