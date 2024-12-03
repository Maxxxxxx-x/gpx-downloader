-- name: GetFileById :one
SELECT * FROM Files WHERE Id = $1 LIMIT 1;

-- name: GetFileByName :one
SELECT * FROM Files WHERE FileName = $1 LIMIT 1;

-- name: InsertFile :one
INSERT INTO Files (
    Id, FileName, SHA512Sum
) VALUES ( $1, $2, $3 ) RETURNING *;

-- name: BulkInsertFiles :copyfrom
INSERT INTO Files ( Id, FileName, SHA512Sum ) VALUES( $1, $2, $3 );

-- name: DeleteFileById :exec
DELETE FROM Files WHERE Id = $1;

-- name: DeleteFileByName :exec
DELETE FROM Files WHERE FileName = $1;

-- name: DropFiles :exec
DELETE FROM Files;
