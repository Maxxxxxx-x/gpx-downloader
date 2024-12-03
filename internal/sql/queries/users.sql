-- name: GetUserById :one
SELECT * FROM Users WHERE Id = $1 LIMIT 1;

-- name: InsertUser :exec
INSERT INTO Users (Id) VALUES ( $1 ) ON CONFLICT DO NOTHING;

-- name: DeleteUser :exec
DELETE FROM Users WHERE Id = $1;

-- name: DropUsers :exec
DELETE FROM Users;
