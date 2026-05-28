-- name: UpsertDomain :one
INSERT INTO domains (host, created_at)
VALUES (?, ?)
ON CONFLICT (host) DO UPDATE SET
    host = excluded.host
RETURNING *;

-- name: GetDomainByHost :one
SELECT * FROM domains
WHERE host = ?
LIMIT 1;

-- name: GetAllDomains :many
SELECT * FROM domains
ORDER BY host;