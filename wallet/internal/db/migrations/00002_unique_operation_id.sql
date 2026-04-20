-- +goose Up
CREATE UNIQUE INDEX transactions_operation_id_key ON transactions (operation_id);

-- +goose Down
DROP INDEX transactions_operation_id_key;
