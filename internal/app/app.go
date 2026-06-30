package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/zerebos/hindsight/internal/browser"
	"github.com/zerebos/hindsight/internal/config"
	"github.com/zerebos/hindsight/internal/db"
	dbgen "github.com/zerebos/hindsight/internal/db/generated"
	"github.com/zerebos/hindsight/internal/ingestion"
)

// App is the central application object. It owns all initialized state and
// exposes the full business logic surface used by both the CLI and the Wails
// binding layer. Neither CLI commands nor Wails bindings contain logic —
// they parse input, call App methods, and format output.
type App struct {
	database  *sql.DB
	queries   *dbgen.Queries
	syncer    *ingestion.Syncer
	settings  config.Settings
	configDir string
	dataDir   string

	// onSettingsChanged, if set, is invoked after settings are successfully
	// persisted. It lets the desktop layer react to changes (toggling
	// launch-at-login, rescheduling the sync timer, …) without the core App
	// depending on any GUI framework. nil in the CLI, where settings are
	// read once per command and never change at runtime.
	onSettingsChanged func(old, updated config.Settings)
}

// New initializes the application: resolves directories, loads config,
// opens the database, runs pending migrations, and prepares the syncer.
// Call Close() when done.
func New() (*App, error) {
	configDir, err := config.ConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve config dir: %w", err)
	}

	dataDir, err := config.DataDir()
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}

	settings, err := config.Load(config.ConfigPath(configDir))
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	database, err := db.Open(config.DBPath(dataDir))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	cacheDir := config.BrowserCacheDir(dataDir)
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		database.Close()
		return nil, fmt.Errorf("create browser cache dir: %w", err)
	}

	return &App{
		database:  database,
		queries:   dbgen.New(database),
		syncer:    ingestion.NewSyncer(database, cacheDir),
		settings:  settings,
		configDir: configDir,
		dataDir:   dataDir,
	}, nil
}

// Close releases all resources held by the App.
func (a *App) Close() error {
	return a.database.Close()
}

// Dirs returns the config and data directories for diagnostic output.
func (a *App) Dirs() (configDir, dataDir string) {
	return a.configDir, a.dataDir
}

// ----------------------------------------------------------------
// Sources
// ----------------------------------------------------------------

// DiscoverSources scans the current machine for installed browsers and
// returns all detected profiles. Does not write to the database.
func (a *App) DiscoverSources() ([]browser.DetectedSource, []error) {
	return browser.Discover()
}

// RegisterSource upserts a detected source into the database and returns
// the saved record. Safe to call repeatedly — existing sources are updated,
// not duplicated.
func (a *App) RegisterSource(ctx context.Context, d browser.DetectedSource) (dbgen.Source, error) {
	return a.queries.UpsertSource(ctx, dbgen.UpsertSourceParams{
		Browser:   d.Browser,
		Profile:   d.Profile,
		Path:      d.Path,
		Label:     sql.NullString{String: d.Label, Valid: true},
		CreatedAt: time.Now().UnixMilli(),
	})
}

// GetSources returns all registered sources ordered by browser and profile.
func (a *App) GetSources(ctx context.Context) ([]dbgen.Source, error) {
	return a.queries.GetAllSources(ctx)
}

// SyncAll runs an incremental sync across all registered sources concurrently.
func (a *App) SyncAll(ctx context.Context) []ingestion.SyncResult {
	return a.syncer.SyncAll(ctx)
}

// SyncSource runs an incremental sync for a single source by ID.
func (a *App) SyncSource(ctx context.Context, sourceID int64) (ingestion.SyncResult, error) {
	sources, err := a.queries.GetAllSources(ctx)
	if err != nil {
		return ingestion.SyncResult{}, fmt.Errorf("get sources: %w", err)
	}

	for _, s := range sources {
		if s.ID == sourceID {
			return a.syncer.SyncSource(ctx, s), nil
		}
	}

	return ingestion.SyncResult{}, fmt.Errorf("source %d not found", sourceID)
}

// DiscoverAndRegister is a convenience method that runs discovery and
// registers all found sources in one call. Returns the registered sources
// and any discovery warnings. Used by the onboarding flow.
func (a *App) DiscoverAndRegister(ctx context.Context) ([]dbgen.Source, []error) {
	detected, errs := browser.Discover()

	var registered []dbgen.Source
	for _, d := range detected {
		src, err := a.RegisterSource(ctx, d)
		if err != nil {
			errs = append(errs, fmt.Errorf("register %q: %w", d.Label, err))
			continue
		}
		registered = append(registered, src)
	}

	return registered, errs
}

// ----------------------------------------------------------------
// Dashboard
// ----------------------------------------------------------------

// DashboardFilter holds the common time range and source filter
// used across all dashboard queries. Zero values mean no filter.
type DashboardFilter struct {
	StartTime int64 // unix ms, 0 = no lower bound
	EndTime   int64 // unix ms, 0 = no upper bound
}

// GetDashboardStats returns the summary stat strip values.
func (a *App) GetDashboardStats(ctx context.Context, f DashboardFilter) (dbgen.GetDashboardStatsRow, error) {
	return a.queries.GetDashboardStats(ctx, dbgen.GetDashboardStatsParams{
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	})
}

// GetTopDomains returns the most visited domains within the filter range.
func (a *App) GetTopDomains(ctx context.Context, f DashboardFilter, limit int64) ([]dbgen.GetTopDomainsRow, error) {
	return a.queries.GetTopDomains(ctx, dbgen.GetTopDomainsParams{
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
		Limit:     limit,
	})
}

// GetVisitTimeSeries returns daily visit counts bucketed by day (unix ms).
func (a *App) GetVisitTimeSeries(ctx context.Context, f DashboardFilter) ([]dbgen.GetVisitTimeSeriesRow, error) {
	return a.queries.GetVisitTimeSeries(ctx, dbgen.GetVisitTimeSeriesParams{
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	})
}

// GetActivityHeatmap returns visit counts bucketed by day-of-week and hour
// in the given timezone. Pass nil to use time.Local.
func (a *App) GetActivityHeatmap(ctx context.Context, f DashboardFilter, loc *time.Location) ([]db.HeatmapCell, error) {
	return db.FetchHeatmap(ctx, a.queries, dbgen.GetRawVisitsForHeatmapParams{
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}, loc)
}

// GetBrowserBreakdown returns visit counts per source ordered by volume.
func (a *App) GetBrowserBreakdown(ctx context.Context, f DashboardFilter) ([]dbgen.GetBrowserBreakdownRow, error) {
	return a.queries.GetBrowserBreakdown(ctx, dbgen.GetBrowserBreakdownParams{
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	})
}

// ----------------------------------------------------------------
// Search
// ----------------------------------------------------------------

// SearchParams holds all search and filter parameters for the query view.
type SearchParams struct {
	Text      string // full-text search across URL and title
	Domain    string // exact domain filter
	StartTime int64  // unix ms, 0 = no lower bound
	EndTime   int64  // unix ms, 0 = no upper bound
	SortBy    string // "recent" (default) | "visits"
	Page      int64  // 0-indexed
	PageSize  int64  // defaults to 50 if 0
}

// SearchResults holds a page of search results with total count for pagination.
type SearchResults struct {
	Visits   []dbgen.SearchVisitsRow
	Total    int64
	Page     int64
	PageSize int64
}

// SearchVisits executes a paginated search query and returns a results page.
func (a *App) SearchVisits(ctx context.Context, p SearchParams) (SearchResults, error) {
	pageSize := p.PageSize
	if pageSize == 0 {
		pageSize = 50
	}

	total, err := a.queries.CountSearchVisits(ctx, dbgen.CountSearchVisitsParams{
		Text:      p.Text,
		Domain:    p.Domain,
		StartTime: p.StartTime,
		EndTime:   p.EndTime,
	})
	if err != nil {
		return SearchResults{}, fmt.Errorf("count search visits: %w", err)
	}

	visits, err := a.queries.SearchVisits(ctx, dbgen.SearchVisitsParams{
		Text:      p.Text,
		Domain:    p.Domain,
		StartTime: p.StartTime,
		EndTime:   p.EndTime,
		Limit:     pageSize,
		Offset:    p.Page * pageSize,
	})
	if err != nil {
		return SearchResults{}, fmt.Errorf("search visits: %w", err)
	}

	return SearchResults{
		Visits:   visits,
		Total:    total,
		Page:     p.Page,
		PageSize: pageSize,
	}, nil
}

// ----------------------------------------------------------------
// Settings
// ----------------------------------------------------------------

// GetSettings returns the current application settings.
func (a *App) GetSettings() config.Settings {
	return a.settings
}

// SetSettingsChangeHandler registers a callback invoked whenever settings are
// updated via UpdateSettings. The handler receives the previous and the newly
// applied settings so it can act only on the fields that actually changed.
// Pass nil to clear. Only the desktop app uses this.
func (a *App) SetSettingsChangeHandler(fn func(old, updated config.Settings)) {
	a.onSettingsChanged = fn
}

// UpdateSettings persists new settings to disk and updates the in-memory state.
// On success it notifies the registered settings-change handler, if any.
func (a *App) UpdateSettings(settings config.Settings) error {
	old := a.settings
	if err := config.Save(config.ConfigPath(a.configDir), settings); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	a.settings = settings
	if a.onSettingsChanged != nil {
		a.onSettingsChanged(old, settings)
	}
	return nil
}

// ----------------------------------------------------------------
// DB Info
// ----------------------------------------------------------------

// DBInfo holds diagnostic information about the database.
type DBInfo struct {
	Path          string
	SizeBytes     int64
	TotalVisits   int64
	TotalDomains  int64
	TotalSources  int64
	SchemaVersion string // most recently applied migration name
}

// GetDBInfo returns diagnostic information about the internal database.
func (a *App) GetDBInfo(ctx context.Context) (DBInfo, error) {
	info := DBInfo{Path: config.DBPath(a.dataDir)}

	// File size
	if fi, err := os.Stat(info.Path); err == nil {
		info.SizeBytes = fi.Size()
	}

	// Row counts
	visits, err := a.queries.CountVisits(ctx)
	if err != nil {
		return info, fmt.Errorf("count visits: %w", err)
	}
	info.TotalVisits = visits

	domains, err := a.queries.GetAllDomains(ctx)
	if err != nil {
		return info, fmt.Errorf("count domains: %w", err)
	}
	info.TotalDomains = int64(len(domains))

	sources, err := a.queries.GetAllSources(ctx)
	if err != nil {
		return info, fmt.Errorf("count sources: %w", err)
	}
	info.TotalSources = int64(len(sources))

	// Latest applied migration
	row := a.database.QueryRowContext(ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`)
	_ = row.Scan(&info.SchemaVersion) // ignore error — table may be empty on fresh DB

	return info, nil
}

// RemoveSource deletes a source from the database by ID.
// Note: existing visits from this source are preserved — they remain in the
// visits table but are orphaned from their source. A future migration could
// add cascade delete behavior if desired.
func (a *App) RemoveSource(ctx context.Context, sourceID int64) error {
	return a.queries.DeleteSource(ctx, sourceID)
}