-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS Files(
    Id TEXT PRIMARY KEY,
    FileName TEXT UNIQUE NOT NULL,
    SHA512Sum TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS files_filename ON Files(FileName);
CREATE INDEX IF NOT EXISTS files_checksum ON Files(SHA512Sum);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Files;
-- +goose StatementEnd

