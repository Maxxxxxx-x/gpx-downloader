-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS Users(
    Id TEXT PRIMARY KEY
);

CREATE INDEX IF NOT EXISTS users_id_idx ON Users(Id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Users;
-- +goose StatementEnd
