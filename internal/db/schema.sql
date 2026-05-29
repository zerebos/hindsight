-- Hindsight schema reference
-- THIS FILE IS NOT EXECUTED DIRECTLY.
-- It reflects the current schema after all migrations have been applied.
-- The source of truth is internal/db/migrations/*.sql
-- Update this file whenever a new migration is added.

CREATE TABLE schema_migrations (
    version    TEXT    NOT NULL PRIMARY KEY,
    applied_at INTEGER NOT NULL
);

CREATE TABLE sources (
    id              INTEGER PRIMARY KEY,
    browser         TEXT    NOT NULL,        -- 'chrome' | 'firefox' | 'safari' etc.
    profile         TEXT    NOT NULL,        -- profile directory name or 'default'
    path            TEXT    NOT NULL UNIQUE, -- path to source DB or import file
    label           TEXT,                    -- user-facing name e.g. "Work Chrome"
    last_synced_at  INTEGER,                 -- wall clock unix ms of last sync run, for UI display
    last_visit_seen INTEGER,                 -- unix ms of newest visit ingested, incremental checkpoint
    last_error      TEXT,                    -- NULL if last sync succeeded
    last_error_at   INTEGER,                 -- unix ms, NULL if no error
    created_at      INTEGER NOT NULL
);

CREATE TABLE domains (
    id          INTEGER PRIMARY KEY,
    host        TEXT    NOT NULL UNIQUE,  -- e.g. "github.com", www. stripped
    favicon_url TEXT,                     -- populated lazily, nullable
    created_at  INTEGER NOT NULL
);

CREATE TABLE visits (
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

-- Stubbed for v2 AI tagging, empty in v1
CREATE TABLE tags (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE,          -- e.g. "tech", "news", "social"
    color      TEXT,                             -- hex color for UI display
    source     TEXT    NOT NULL DEFAULT 'user',  -- 'user' | 'ai'
    created_at INTEGER NOT NULL
);

-- Stubbed for v2 AI tagging, empty in v1
-- confidence is NULL for user tags, 0.0-1.0 for AI-assigned tags
CREATE TABLE domain_tags (
    domain_id  INTEGER NOT NULL REFERENCES domains(id),
    tag_id     INTEGER NOT NULL REFERENCES tags(id),
    confidence REAL,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (domain_id, tag_id)
);

CREATE INDEX idx_visits_visited_at ON visits(visited_at);
CREATE INDEX idx_visits_domain_id  ON visits(domain_id);
CREATE INDEX idx_visits_source_id  ON visits(source_id);
CREATE INDEX idx_visits_url        ON visits(url);
CREATE INDEX idx_domains_host      ON domains(host);

CREATE UNIQUE INDEX idx_visits_dedup
    ON visits(url, visited_at, source_id);