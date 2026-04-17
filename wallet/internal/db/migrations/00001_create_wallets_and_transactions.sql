-- +goose Up
CREATE TABLE wallets (
    id          UUID        PRIMARY KEY,
    user_id     UUID        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    balance     NUMERIC     NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE transactions (
    id           UUID        PRIMARY KEY,
    wallet_id    UUID        NOT NULL REFERENCES wallets(id),
    value        NUMERIC     NOT NULL,
    description  TEXT        NOT NULL DEFAULT '',
    operation_id UUID        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE transactions;
DROP TABLE wallets;
