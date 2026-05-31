package app

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

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
