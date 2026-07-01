# Hindsight

Personal browser history intelligence. Hindsight reads your browser history directly from local browser databases, normalizes it into a unified store, and gives you a powerful analytics dashboard and search interface across all your browsers and profiles.

> **Status:** Active development, CLI and ingestion pipeline complete, desktop app (Wails v3) in progress.

---

## Features

**Analytics dashboard**

Summary and charts:
- Headline stats: total visits, unique domains, active days, total duration, and sources
- Interactive visit-history chart (area/line with axis ticks, gridlines, and hover crosshair)
- Activity heatmap by day of week and hour of day (timezone-aware), with an intensity legend and peak highlight
- Busiest-hours and busiest-days distributions
- Top visited domains with visit counts, and a per-browser breakdown donut

Derived insights (computed from your data, no extra queries):
- Daily average, busiest day, peak hour, top weekday, longest streak, chronotype, weekend share, top-domain focus, and momentum

Deeper stats:
- **Period-over-period trends** - visits/domains/active-days change versus the immediately preceding equal window, plus domain rank movers
- **Tracking & privacy** - share of visits that carried tracking parameters, the most common parameters, and the most-tracked domains (derived from the preserved `raw_url`)
- **Domain depth & diversity** - domains newly discovered this period, one-off (visited-once) domains, and top-10 concentration
- **Time spent** - top domains by time, average per timed visit, and duration coverage (for browsers that report it)

**Search**
- Full-text search across URLs and page titles
- Filter by domain, date range, and browser source
- Paginated results

**Browser support**

| Browser | Engine | Windows | macOS | Linux |
|---|---|---|---|---|
| Chrome | Chromium | ✓ | ✓ | ✓ |
| Edge | Chromium | ✓ | ✓ | ✓ |
| Brave | Chromium | ✓ | ✓ | ✓ |
| Vivaldi | Chromium | ✓ | ✓ | ✓ |
| Opera | Chromium | ✓ | ✓ | ✓ |
| Arc | Chromium | — | ✓ | — |
| Helium | Chromium | ✓ | — | — |
| Firefox | Gecko | ✓ | ✓ | ✓ |
| Zen | Gecko | ✓ | ✓ | ✓ |
| LibreWolf | Gecko | ✓ | ✓ | ✓ |
| Floorp | Gecko | ✓ | ✓ | ✓ |
| Safari | WebKit | — | ✓ | — |

Multiple profiles per browser are supported. History databases are read safely via a copy-then-read pattern, your browser databases are never modified.

**Incremental sync**: subsequent syncs only read new history entries using per-source checkpointing. First sync may take a few seconds for large histories; incremental syncs are fast.

---

## Data storage

Hindsight stores all data locally in a SQLite database. No data leaves your machine.

| Platform | Config | Database |
|---|---|---|
| Windows | `%APPDATA%\hindsight\` | `%APPDATA%\hindsight\` |
| macOS | `~/Library/Application Support/hindsight/` | `~/Library/Application Support/hindsight/` |
| Linux | `~/.config/hindsight/` | `~/.local/share/hindsight/` |

**URL normalization**: known tracking parameters (`utm_*`, `fbclid`, `gclid`, etc.) are stripped from stored URLs. The original URL is preserved in `raw_url` for future analytics.

---

## CLI

A development CLI is available alongside the desktop app.

```
hindsight discover              Scan for installed browsers (dry run)
hindsight discover --register   Register found sources in the database
hindsight sync                  Sync all registered sources
hindsight sync --source <id>    Sync a specific source by ID
hindsight stats                 Print all-time dashboard summary
hindsight stats --days 30       Print summary for the last 30 days
hindsight search <query>        Search visit history
hindsight search --domain github.com --days 7
hindsight sources list          List registered sources with sync status
hindsight sources remove <id>   Remove a source (visits are preserved)
hindsight info                  Show database stats and configuration
```

---

## Development

### Prerequisites

- Go 1.22+
- [sqlc](https://sqlc.dev) - for regenerating database query code
- [Wails v3](https://v3.wails.io) - for building the desktop app
- Node.js - for the frontend

### Project structure

```
hindsight/
├── main.go                     Wails desktop entry point
├── services.go                 Wails service bindings (wrap internal/app)
├── tray.go                     System tray setup
├── cmd/hindsight/
│   └── main.go                 CLI entry point (Cobra)
└── internal/
    ├── app/
    │   └── app.go              Core business logic (used by CLI + Wails)
    ├── browser/
    │   ├── browser.go          Browser family types, Discover()
    │   ├── chromium.go         Chromium variant discovery
    │   ├── firefox.go          Firefox variant discovery + profiles.ini parsing
    │   ├── safari.go           Safari discovery
    │   └── paths.go            OS-aware profile path resolution
    ├── config/
    │   └── config.go           Settings, platform directory resolution
    ├── db/
    │   ├── db.go               Database open + configuration
    │   ├── migrate.go          Migration runner (embedded SQL files)
    │   ├── heatmap.go          Timezone-aware heatmap bucketing
    │   ├── schema.sql          Current schema reference (not executed directly)
    │   ├── migrations/         Numbered SQL migration files
    │   ├── queries/            sqlc input query files
    │   └── generated/          sqlc generated code (do not edit)
    └── ingestion/
        ├── ingestion.go        Syncer, SyncAll, SyncSource
        ├── copy.go             Safe copy-then-read for browser databases
        ├── normalize.go        URL normalization, timestamp conversions
        ├── chromium.go         Chromium history database reader
        ├── firefox.go          Firefox places.sqlite reader
        └── safari.go           Safari History.db reader
```

### Running the CLI

```bash
go run ./cmd/hindsight
```

### Building the desktop app

```bash
wails3 build
# or for development with hot reload:
wails3 dev
```

### Regenerating database code

After modifying files in `internal/db/queries/`:

```bash
sqlc generate
```

### Adding a new browser

**Chromium-based:** add one entry to `knownChromiumVariants` in `internal/browser/chromium.go` and one entry to `browserFamilies` in `internal/browser/browser.go`.

**Firefox-based:** add one entry to `knownFirefoxVariants` in `internal/browser/firefox.go` and one entry to `browserFamilies` in `internal/browser/browser.go`.

### Adding a schema migration

1. Create `internal/db/migrations/NNNN_description.sql`
2. Update `internal/db/schema.sql` to reflect the new state
3. Run `sqlc generate` if queries were affected

Migrations run automatically on app startup. During development, delete the database file to start fresh.

---

## Architecture notes

**Visit count semantics**: the `visit_count` column on the `visits` table always reflects individual visit rows (value is always 1 for current browser sources). Use `SUM(visit_count)` for totals, never `COUNT(*)`. This convention is documented in schema comments and is designed to accommodate future sources that may report aggregated counts.

**Timestamp storage**: all timestamps are stored as Unix milliseconds (INTEGER). Browser-native timestamp formats (Chromium: microseconds since 1601, Firefox: microseconds since 1970, Safari: seconds since 2001) are converted during ingestion. Conversion functions live in `internal/ingestion/normalize.go`.

**Heatmap timezone handling**: day-of-week and hour-of-day bucketing is done in Go using `time.In(loc)` rather than SQL. This correctly handles DST transitions that a fixed UTC offset cannot. The heatmap reflects the user's current system timezone applied uniformly across all historical data. This is intentional, as the heatmap answers "when in my day do I browse" which is a present-tense lifestyle question.

**URL normalization**: tracking parameters are stripped at ingestion time. The original URL is preserved in `raw_url` (nullable, only set when normalization changed the URL) for future analytics use cases such as tracking parameter frequency analysis.

---

## Roadmap

**v1: Desktop app**
- [ ] Wails v3 desktop shell
- [x] Analytics dashboard UI
- [ ] Search UI
- [ ] Settings UI with source management
- [ ] Sync scheduling (launch, wake, interval)

**v2**
- [ ] AI-powered domain tagging (local model preferred)
- [ ] Browser extension + live sync
- [ ] Self-hosted web version
- [x] Tracking parameter analytics

**Future**
- [ ] Cross-device sync
- [ ] Full page content indexing
- [ ] CDN/subdomain filtering rules

---

## License

TBD