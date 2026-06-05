-- NOTE: sqlc's SQLite parser does not handle non-ASCII characters in comments.
-- Use plain ASCII only - avoid em dashes, smart quotes, etc.

-- name: InsertVisit :execresult
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
) VALUES (
    @url,
    @raw_url,
    @title,
    @domain_id,
    @source_id,
    @visited_at,
    @duration_ms,
    @visit_count,
    @created_at
)
ON CONFLICT (url, visited_at, source_id) DO NOTHING;

-- name: CountVisits :one
SELECT COUNT(*) FROM visits;

-- name: CountVisitsBySource :one
SELECT COUNT(*) FROM visits
WHERE source_id = @source_id;

-- name: GetLatestVisitTime :one
SELECT CAST(COALESCE(MAX(visited_at), 0) AS INTEGER) FROM visits
WHERE source_id = @source_id;

-- name: GetTopDomains :many
SELECT
    d.host,
    CAST(SUM(v.visit_count) AS INTEGER) AS total_visits
FROM visits v
JOIN domains d ON v.domain_id = d.id
WHERE
    (CAST(@start_time AS INTEGER) = 0 OR v.visited_at >= CAST(@start_time AS INTEGER)) AND
    (CAST(@end_time   AS INTEGER) = 0 OR v.visited_at <= CAST(@end_time   AS INTEGER))
GROUP BY d.id, d.host
ORDER BY total_visits DESC
LIMIT @limit;

-- name: GetVisitTimeSeries :many
-- Returns daily visit counts bucketed by day (unix ms at midnight UTC).
SELECT
    (visited_at / 86400000) * 86400000 AS day,
    CAST(SUM(visit_count) AS INTEGER) AS total_visits
FROM visits
WHERE
    (CAST(@start_time AS INTEGER) = 0 OR visited_at >= CAST(@start_time AS INTEGER)) AND
    (CAST(@end_time   AS INTEGER) = 0 OR visited_at <= CAST(@end_time   AS INTEGER))
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
    v.duration_ms,
    d.host AS domain
FROM visits v
JOIN domains d ON v.domain_id = d.id
WHERE
    (CAST(@text   AS TEXT) = '' OR v.url   LIKE '%' || CAST(@text AS TEXT) || '%'
                                OR v.title LIKE '%' || CAST(@text AS TEXT) || '%') AND
    (CAST(@domain AS TEXT) = '' OR d.host  = CAST(@domain AS TEXT)) AND
    (CAST(@start_time AS INTEGER) = 0 OR v.visited_at >= CAST(@start_time AS INTEGER)) AND
    (CAST(@end_time   AS INTEGER) = 0 OR v.visited_at <= CAST(@end_time   AS INTEGER))
ORDER BY v.visited_at DESC
LIMIT  @limit
OFFSET @offset;

-- name: CountSearchVisits :one
SELECT COUNT(*)
FROM visits v
JOIN domains d ON v.domain_id = d.id
WHERE
    (CAST(@text   AS TEXT) = '' OR v.url   LIKE '%' || CAST(@text AS TEXT) || '%'
                                OR v.title LIKE '%' || CAST(@text AS TEXT) || '%') AND
    (CAST(@domain AS TEXT) = '' OR d.host  = CAST(@domain AS TEXT)) AND
    (CAST(@start_time AS INTEGER) = 0 OR v.visited_at >= CAST(@start_time AS INTEGER)) AND
    (CAST(@end_time   AS INTEGER) = 0 OR v.visited_at <= CAST(@end_time   AS INTEGER));

-- name: GetRawVisitsForHeatmap :many
-- Returns raw visited_at timestamps and visit counts for heatmap bucketing.
-- Timezone-aware day/hour extraction is done in Go using time.In(loc) so
-- that sqlc does not need to handle named parameters in arithmetic expressions
-- (a known parser limitation). This also makes timezone handling fully testable.
SELECT
    visited_at,
    visit_count
FROM visits
WHERE
    (CAST(@start_time AS INTEGER) = 0 OR visited_at >= CAST(@start_time AS INTEGER)) AND
    (CAST(@end_time   AS INTEGER) = 0 OR visited_at <= CAST(@end_time   AS INTEGER));

-- name: GetDashboardStats :one
-- Returns a single-row summary for the dashboard stat strip.
-- active_days is computed in UTC - close enough for a summary stat and
-- avoids timezone arithmetic in SQL entirely.
SELECT
    CAST(SUM(visit_count) AS INTEGER) AS total_visits,
    COUNT(DISTINCT domain_id) AS unique_domains,
    COUNT(DISTINCT (visited_at / 86400000)) AS active_days,
    CAST(SUM(duration_ms) AS INTEGER) AS total_duration_ms
FROM visits
WHERE
    (CAST(@start_time AS INTEGER) = 0 OR visited_at >= CAST(@start_time AS INTEGER)) AND
    (CAST(@end_time   AS INTEGER) = 0 OR visited_at <= CAST(@end_time   AS INTEGER));