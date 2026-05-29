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
    SUM(v.visit_count) AS total_visits
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
    SUM(visit_count) AS total_visits
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