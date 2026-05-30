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
SET last_synced_at  = @last_synced_at,   -- wall clock time of this sync run
    last_visit_seen = @last_visit_seen,   -- newest visit timestamp ingested
    last_error      = NULL,
    last_error_at   = NULL
WHERE id = @id;

-- name: UpdateSourceSyncError :exec
UPDATE sources
SET last_error    = @last_error,
    last_error_at = @last_error_at
WHERE id = @id;

-- name: DeleteSource :exec
DELETE FROM sources WHERE id = @id;

-- name: GetBrowserBreakdown :many
SELECT
    s.id,
    s.browser,
    s.label,
    SUM(v.visit_count) AS total_visits
FROM sources s
JOIN visits v ON v.source_id = s.id
WHERE
    (CAST(@start_time AS INTEGER) = 0 OR v.visited_at >= CAST(@start_time AS INTEGER)) AND
    (CAST(@end_time   AS INTEGER) = 0 OR v.visited_at <= CAST(@end_time   AS INTEGER))
GROUP BY s.id
ORDER BY total_visits DESC;