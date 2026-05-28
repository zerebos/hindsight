-- name: UpsertSource :one
INSERT INTO sources (browser, profile, path, label, created_at)
VALUES (@browser, @profile, @path, @label, @created_at)
ON CONFLICT (path) DO UPDATE SET
    label   = excluded.label,
    browser = excluded.browser,
    profile = excluded.profile
RETURNING *;

-- name: GetSourceByPath :one
SELECT * FROM sources
WHERE path = @path
LIMIT 1;

-- name: GetAllSources :many
SELECT * FROM sources
ORDER BY browser, profile;

-- name: UpdateSourceSyncSuccess :exec
UPDATE sources
SET last_synced_at = @last_synced_at,
    last_error     = NULL,
    last_error_at  = NULL
WHERE id = @id;

-- name: UpdateSourceSyncError :exec
UPDATE sources
SET last_error    = @last_error,
    last_error_at = @last_error_at
WHERE id = @id;

-- name: DeleteSource :exec
DELETE FROM sources WHERE id = @id;