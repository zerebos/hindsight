-- Migration: 000002_split_last_synced_at.sql
-- Splits the overloaded last_synced_at column on sources into two columns
-- with distinct, unambiguous semantics:
--
--   last_synced_at  = wall clock unix ms of when the sync job last ran
--                     used for UI display ("last checked X minutes ago")
--
--   last_visit_seen = unix ms of the newest visit row ingested from this source
--                     used as the WHERE clause lower bound for incremental sync
--
-- SQLite does not support DROP COLUMN or RENAME COLUMN on tables with
-- foreign key references in older versions, but ADD COLUMN is always safe.
-- We add last_visit_seen as a new column and leave last_synced_at in place
-- with its semantics redefined to wall-clock-only going forward.

ALTER TABLE sources ADD COLUMN last_visit_seen INTEGER;