-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS Records(
    Id TEXT PRIMARY KEY,
    UserId TEXT NOT NULL,
    FileId TEXT UNIQUE NOT NULL,
    Duration FLOAT8,
    Distance FLOAT8,
    Ascent FLOAT8,
    Descent FLOAT8,
    ElevationDiff FLOAT8,
    Trails TEXT,
    RawData TEXT
);


CREATE INDEX IF NOT EXISTS records_userid_idx ON Records(UserId);
CREATE INDEX IF NOT EXISTS records_fileid_idx ON Records(FileId);
CREATE INDEX IF NOT EXISTS records_trails_idx ON Records(Trails);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Record;
-- +goose StatementEnd
