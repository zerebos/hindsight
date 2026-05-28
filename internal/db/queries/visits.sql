-- name: InsertVisit :exec
INSERT INTO visits (
    url,
    raw_url,
    title,
    domain_id,
    source_id,
    visited_at,
    duration_ms,
    visit_count,
    created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (url, visited_at, source_id) DO NOTHING;

-- name: CountVisits :one
SELECT COUNT(*) FROM visits;

-- name: CountVisitsBySource :one
SELECT COUNT(*) FROM visits
WHERE source_id = ?;

-- name: GetLatestVisitTime :one
SELECT COALESCE(MAX(visited_at), 0) FROM visits
WHERE source_id = ?;

-- name: GetTopDomains :many
SELECT
    d.host,
    SUM(v.visit_count) AS total_visits
FROM visits v
JOIN domains d ON v.domain_id = d.id
WHERE
    (? = 0 OR v.visited_at >= ?) AND
    (? = 0 OR v.visited_at <= ?)
GROUP BY d.id, d.host
ORDER BY total_visits DESC
LIMIT ?;

-- name: GetVisitTimeSeries :many
-- Returns daily visit counts bucketed by day (unix ms at midnight UTC).
-- Cast to integer division truncates to day boundary.
SELECT
    (visited_at / 86400000) * 86400000 AS day,
    SUM(visit_count) AS total_visits
FROM visits
WHERE
    (? = 0 OR visited_at >= ?) AND
    (? = 0 OR visited_at <= ?)
GROUP BY day
ORDER BY day ASC;

-- name: SearchVisits :many
SELECT
    v.id,
    v.url,
    v.title,
    v.visited_at,
    v.visit_count,
    v.source_id,
    d.host AS domain
FROM visits v
JOIN domains d ON v.domain_id = d.id
WHERE
    (? = '' OR v.url LIKE '%' || ? || '%' OR v.title LIKE '%' || ? || '%') AND
    (? = '' OR d.host = ?) AND
    (? = 0  OR v.visited_at >= ?) AND
    (? = 0  OR v.visited_at <= ?)
ORDER BY v.visited_at DESC
LIMIT  ?
OFFSET ?;

-- name: CountSearchVisits :one
SELECT COUNT(*)
FROM visits v
JOIN domains d ON v.domain_id = d.id
WHERE
    (? = '' OR v.url LIKE '%' || ? || '%' OR v.title LIKE '%' || ? || '%') AND
    (? = '' OR d.host = ?) AND
    (? = 0  OR v.visited_at >= ?) AND
    (? = 0  OR v.visited_at <= ?);