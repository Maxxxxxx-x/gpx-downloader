-- name: GetRecordById :one
SELECT * FROM Records WHERE Id = $1 LIMIT 1;

-- name: GetRecordByFileId :one
SELECT * FROM Records WHERE FileId = $1 LIMIT 1;

-- name: GetRecordsByUserId :many
SELECT * FROM Records WHERE UserId = $1;

-- name: GetRecordsByTrail :many
SELECT * FROM Records WHERE Trails = $1;

-- name: GetRecordsOfUserOnTrail :many
SELECT * FROM Records WHERE UserId = $1 AND Trails = $2;

-- name: InsertRecord :one
INSERT INTO Records (
    Id, UserId, FileId, Duration, Distance, Ascent, Descent, ElevationDiff, Trails, RawData
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: BulkInsertRecord :copyfrom
INSERT INTO Records (
    Id, UserId, FileId, Duration, Distance, Ascent, Descent, ElevationDiff, Trails, RawData
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: DeleteRecordById :exec
DELETE FROM Records WHERE Id = $1;

-- name: DeleteRecordsByUserId :exec
DELETE FROM Records WHERE UserId = $1;

-- name: DeleteRecordByFileId :exec
DELETE FROM Records WHERE FileId = $1;

-- name: DropRecords :exec
DELETE FROM Records;
