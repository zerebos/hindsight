-- Migration: 000001_initial_schema.sql
-- Initial Hindsight schema.
-- All timestamps stored as unix milliseconds (INTEGER).
-- Always use SUM(visit_count) for total visit counts, never COUNT(*).

CREATE TABLE sources (
    id              INTEGER PRIMARY KEY,
    browser         TEXT    NOT NULL,
    profile         TEXT    NOT NULL,
    path            TEXT    NOT NULL UNIQUE,
    label           TEXT,
    last_synced_at  INTEGER,
    last_error      TEXT,
    last_error_at   INTEGER,
    created_at      INTEGER NOT NULL
);

CREATE TABLE domains (
    id          INTEGER PRIMARY KEY,
    host        TEXT    NOT NULL UNIQUE,
    favicon_url TEXT,
    created_at  INTEGER NOT NULL
);

CREATE TABLE visits (
    id          INTEGER PRIMARY KEY,
    url         TEXT    NOT NULL,
    raw_url     TEXT,
    title       TEXT,
    domain_id   INTEGER NOT NULL REFERENCES domains(id),
    source_id   INTEGER NOT NULL REFERENCES sources(id),
    visited_at  INTEGER NOT NULL,
    duration_ms INTEGER,
    visit_count INTEGER NOT NULL DEFAULT 1,
    created_at  INTEGER NOT NULL
);

-- Stubbed for v2 AI tagging, empty in v1
CREATE TABLE tags (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE,
    color      TEXT,
    source     TEXT    NOT NULL DEFAULT 'user',
    created_at INTEGER NOT NULL
);

-- Stubbed for v2 AI tagging, empty in v1
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