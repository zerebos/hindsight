package app

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/zerebos/hindsight/internal/browser"
	"github.com/zerebos/hindsight/internal/config"
	"github.com/zerebos/hindsight/internal/db"
	dbgen "github.com/zerebos/hindsight/internal/db/generated"
	"github.com/zerebos/hindsight/internal/ingestion"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open(): %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	return &App{
		database:  database,
		queries:   dbgen.New(database),
		syncer:    ingestion.NewSyncer(database, t.TempDir()),
		settings:  config.Defaults(),
		configDir: t.TempDir(),
		dataDir:   t.TempDir(),
	}
}

func seedSearchVisit(t *testing.T, app *App) {
	t.Helper()

	ctx := context.Background()

	domain, err := app.queries.UpsertDomain(ctx, dbgen.UpsertDomainParams{
		Host:      "example.com",
		CreatedAt: 1,
	})
	if err != nil {
		t.Fatalf("UpsertDomain(): %v", err)
	}

	source, err := app.queries.UpsertSource(ctx, dbgen.UpsertSourceParams{
		Browser:   "chrome",
		Profile:   "Default",
		Path:      filepath.Join(t.TempDir(), "History"),
		Label:     sql.NullString{String: "Chrome", Valid: true},
		CreatedAt: 1,
	})
	if err != nil {
		t.Fatalf("UpsertSource(): %v", err)
	}

	_, err = app.queries.InsertVisit(ctx, dbgen.InsertVisitParams{
		Url:        "https://example.com/page",
		RawUrl:     sql.NullString{String: "https://example.com/page?utm_source=newsletter", Valid: true},
		Title:      sql.NullString{String: "Example Page", Valid: true},
		DomainID:   domain.ID,
		SourceID:   source.ID,
		VisitedAt:  1706750280000,
		DurationMs: sql.NullInt64{},
		VisitCount: 1,
		CreatedAt:  1,
	})
	if err != nil {
		t.Fatalf("InsertVisit(): %v", err)
	}
}

func TestSyncSourceNotFound(t *testing.T) {
	app := newTestApp(t)

	_, err := app.SyncSource(context.Background(), 12345)
	if err == nil {
		t.Fatalf("SyncSource() error = nil, want non-nil")
	}
	if got := err.Error(); got != "source 12345 not found" {
		t.Fatalf("SyncSource() error = %q, want %q", got, "source 12345 not found")
	}
}

func TestSearchVisitsDefaultsPageSize(t *testing.T) {
	app := newTestApp(t)
	seedSearchVisit(t, app)

	got, err := app.SearchVisits(context.Background(), SearchParams{})
	if err != nil {
		t.Fatalf("SearchVisits() error = %v", err)
	}

	if got.PageSize != 50 {
		t.Fatalf("PageSize = %d, want 50", got.PageSize)
	}
	if got.Total != 1 {
		t.Fatalf("Total = %d, want 1", got.Total)
	}
	if got.Page != 0 {
		t.Fatalf("Page = %d, want 0", got.Page)
	}
	if len(got.Visits) != 1 {
		t.Fatalf("len(Visits) = %d, want 1", len(got.Visits))
	}
	if got.Visits[0].Url != "https://example.com/page" {
		t.Fatalf("Visits[0].Url = %q, want %q", got.Visits[0].Url, "https://example.com/page")
	}
	if got.Visits[0].Domain != "example.com" {
		t.Fatalf("Visits[0].Domain = %q, want %q", got.Visits[0].Domain, "example.com")
	}
}

func TestRegisterAndGetSources(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	src, err := app.RegisterSource(ctx, browser.DetectedSource{
		Browser: "chrome",
		Profile: "Default",
		Path:    filepath.Join(t.TempDir(), "History"),
		Label:   "Chrome (Default)",
		Family:  browser.FamilyChromium,
	})
	if err != nil {
		t.Fatalf("RegisterSource() error = %v", err)
	}
	if src.Browser != "chrome" || src.Profile != "Default" {
		t.Fatalf("RegisterSource() browser/profile mismatch: %+v", src)
	}

	sources, err := app.GetSources(ctx)
	if err != nil {
		t.Fatalf("GetSources() error = %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("GetSources() len = %d, want 1", len(sources))
	}
	if sources[0].ID != src.ID {
		t.Fatalf("GetSources()[0].ID = %d, want %d", sources[0].ID, src.ID)
	}
}

func TestRemoveSource(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	src, err := app.RegisterSource(ctx, browser.DetectedSource{
		Browser: "firefox",
		Profile: "default-release",
		Path:    filepath.Join(t.TempDir(), "places.sqlite"),
		Label:   "Firefox (default-release)",
		Family:  browser.FamilyFirefox,
	})
	if err != nil {
		t.Fatalf("RegisterSource() error = %v", err)
	}

	if err := app.RemoveSource(ctx, src.ID); err != nil {
		t.Fatalf("RemoveSource() error = %v", err)
	}

	sources, err := app.GetSources(ctx)
	if err != nil {
		t.Fatalf("GetSources() error = %v", err)
	}
	if len(sources) != 0 {
		t.Fatalf("GetSources() after remove len = %d, want 0", len(sources))
	}
}

func TestGetAndUpdateSettings(t *testing.T) {
	app := newTestApp(t)

	got := app.GetSettings()
	if got != config.Defaults() {
		t.Fatalf("GetSettings() = %+v, want defaults", got)
	}

	updated := got
	updated.General.Theme = "dark"
	updated.Sync.IntervalMinutes = 15

	if err := app.UpdateSettings(updated); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	now := app.GetSettings()
	if now.General.Theme != "dark" {
		t.Fatalf("Theme = %q, want %q", now.General.Theme, "dark")
	}
	if now.Sync.IntervalMinutes != 15 {
		t.Fatalf("IntervalMinutes = %d, want 15", now.Sync.IntervalMinutes)
	}
}

func TestGetDBInfo(t *testing.T) {
	app := newTestApp(t)
	seedSearchVisit(t, app)

	info, err := app.GetDBInfo(context.Background())
	if err != nil {
		t.Fatalf("GetDBInfo() error = %v", err)
	}
	if info.TotalVisits != 1 {
		t.Fatalf("TotalVisits = %d, want 1", info.TotalVisits)
	}
	if info.TotalDomains != 1 {
		t.Fatalf("TotalDomains = %d, want 1", info.TotalDomains)
	}
	if info.TotalSources != 1 {
		t.Fatalf("TotalSources = %d, want 1", info.TotalSources)
	}
}

func TestGetDashboardStats(t *testing.T) {
	app := newTestApp(t)
	seedSearchVisit(t, app)

	stats, err := app.GetDashboardStats(context.Background(), DashboardFilter{})
	if err != nil {
		t.Fatalf("GetDashboardStats() error = %v", err)
	}
	if stats.TotalVisits != 1 {
		t.Fatalf("TotalVisits = %d, want 1", stats.TotalVisits)
	}
	if stats.UniqueDomains != 1 {
		t.Fatalf("UniqueDomains = %d, want 1", stats.UniqueDomains)
	}
	if stats.ActiveDays != 1 {
		t.Fatalf("ActiveDays = %d, want 1", stats.ActiveDays)
	}
}

func TestGetTopDomains(t *testing.T) {
	app := newTestApp(t)
	seedSearchVisit(t, app)

	rows, err := app.GetTopDomains(context.Background(), DashboardFilter{}, 10)
	if err != nil {
		t.Fatalf("GetTopDomains() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("GetTopDomains() len = %d, want 1", len(rows))
	}
	if rows[0].Host != "example.com" {
		t.Fatalf("Host = %q, want %q", rows[0].Host, "example.com")
	}
	if rows[0].TotalVisits != 1 {
		t.Fatalf("TotalVisits = %d, want 1", rows[0].TotalVisits)
	}
}

func TestGetVisitTimeSeries(t *testing.T) {
	app := newTestApp(t)
	seedSearchVisit(t, app)

	series, err := app.GetVisitTimeSeries(context.Background(), DashboardFilter{})
	if err != nil {
		t.Fatalf("GetVisitTimeSeries() error = %v", err)
	}
	if len(series) != 1 {
		t.Fatalf("GetVisitTimeSeries() len = %d, want 1", len(series))
	}
	if series[0].TotalVisits != 1 {
		t.Fatalf("series[0].TotalVisits = %d, want 1", series[0].TotalVisits)
	}
}

func TestGetActivityHeatmap(t *testing.T) {
	app := newTestApp(t)
	seedSearchVisit(t, app)

	// seedSearchVisit uses VisitedAt=1706750280000 = 2024-02-01 01:18:00 UTC = Thursday hour 1
	cells, err := app.GetActivityHeatmap(context.Background(), DashboardFilter{}, time.UTC)
	if err != nil {
		t.Fatalf("GetActivityHeatmap() error = %v", err)
	}
	if len(cells) != 1 {
		t.Fatalf("GetActivityHeatmap() len = %d, want 1", len(cells))
	}
	if cells[0].Day != int(time.Thursday) {
		t.Fatalf("cell.Day = %d, want %d (Thursday)", cells[0].Day, int(time.Thursday))
	}
	if cells[0].Hour != 1 {
		t.Fatalf("cell.Hour = %d, want 1", cells[0].Hour)
	}
	if cells[0].TotalVisits != 1 {
		t.Fatalf("cell.TotalVisits = %d, want 1", cells[0].TotalVisits)
	}
}

func TestGetBrowserBreakdown(t *testing.T) {
	app := newTestApp(t)
	seedSearchVisit(t, app)

	rows, err := app.GetBrowserBreakdown(context.Background(), DashboardFilter{})
	if err != nil {
		t.Fatalf("GetBrowserBreakdown() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("GetBrowserBreakdown() len = %d, want 1", len(rows))
	}
	if rows[0].Browser != "chrome" {
		t.Fatalf("Browser = %q, want %q", rows[0].Browser, "chrome")
	}
	if rows[0].TotalVisits != 1 {
		t.Fatalf("TotalVisits = %d, want 1", rows[0].TotalVisits)
	}
}
