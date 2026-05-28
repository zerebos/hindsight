-- Hindsight schema
-- All timestamps are stored as unix milliseconds (INTEGER)
-- Always use SUM(visit_count) for total visit counts, never COUNT(*)
-- visit_count reflects source browser behavior (always 1 for individual visit rows)

CREATE TABLE IF NOT EXISTS sources (
    id             INTEGER PRIMARY KEY,
    browser        TEXT    NOT NULL,  -- 'chrome' | 'firefox' | 'safari'
    profile        TEXT    NOT NULL,  -- profile directory name or 'default'
    path           TEXT    NOT NULL,  -- path to source DB or import file
    label          TEXT,              -- user-facing name e.g. "Work Chrome"
    last_synced_at INTEGER,           -- unix ms, NULL if never synced
    last_error     TEXT,              -- NULL if last sync succeeded
    last_error_at  INTEGER,           -- unix ms, NULL if no error
    created_at     INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS domains (
    id          INTEGER PRIMARY KEY,
    host        TEXT    NOT NULL UNIQUE,  -- e.g. "github.com", www. stripped
    favicon_url TEXT,                     -- populated lazily, nullable
    created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS visits (
    id          INTEGER PRIMARY KEY,
    url         TEXT    NOT NULL,  -- normalized: tracking params stripped
    raw_url     TEXT,              -- original url, only set when normalization changed it
    title       TEXT,
    domain_id   INTEGER NOT NULL REFERENCES domains(id),
    source_id   INTEGER NOT NULL REFERENCES sources(id),
    visited_at  INTEGER NOT NULL,  -- unix ms
    duration_ms INTEGER,           -- nullable, not all browsers report this
    visit_count INTEGER NOT NULL DEFAULT 1,
    created_at  INTEGER NOT NULL
);

-- Stubbed for v2 AI tagging -- empty in v1
CREATE TABLE IF NOT EXISTS tags (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE,  -- e.g. "tech", "news", "social"
    color      TEXT,                     -- hex color for UI display
    source     TEXT    NOT NULL DEFAULT 'user',  -- 'user' | 'ai'
    created_at INTEGER NOT NULL
);

-- Domain<->Tag join -- stubbed for v2, empty in v1
-- confidence is NULL for user tags, 0.0-1.0 for AI-assigned tags
CREATE TABLE IF NOT EXISTS domain_tags (
    domain_id  INTEGER NOT NULL REFERENCES domains(id),
    tag_id     INTEGER NOT NULL REFERENCES tags(id),
    confidence REAL,              -- NULL = user tag, 0.0-1.0 = ai confidence
    created_at INTEGER NOT NULL,
    PRIMARY KEY (domain_id, tag_id)
);

-- Indexes for dashboard queries (hit constantly)
CREATE INDEX IF NOT EXISTS idx_visits_visited_at ON visits(visited_at);
CREATE INDEX IF NOT EXISTS idx_visits_domain_id  ON visits(domain_id);
CREATE INDEX IF NOT EXISTS idx_visits_source_id  ON visits(source_id);

-- Indexes for search queries
CREATE INDEX IF NOT EXISTS idx_visits_url   ON visits(url);
CREATE INDEX IF NOT EXISTS idx_domains_host ON domains(host);

-- Dedup guard: same url + timestamp + source = same visit
CREATE UNIQUE INDEX IF NOT EXISTS idx_visits_dedup
    ON visits(url, visited_at, source_id);