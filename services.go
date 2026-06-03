package main

import (
	"context"
	"time"

	"github.com/zerebos/hindsight/internal/app"
	"github.com/zerebos/hindsight/internal/browser"
	"github.com/zerebos/hindsight/internal/config"
	idb "github.com/zerebos/hindsight/internal/db"
	dbgen "github.com/zerebos/hindsight/internal/db/generated"
	"github.com/zerebos/hindsight/internal/ingestion"
)

// Services are thin wrappers around the App struct that expose business logic
// to the Wails frontend. They do not contain logic themselves.
//
// Service methods do not take context.Context parameters since Wails generates
// the JS/TS binding from the method signature and context is not a JS concept.
// Each service stores a context set during startup for use in App calls.
//
// NOTE: If Wails v3 exposes a ServiceStartup lifecycle hook in your version,
// use it to set svc.ctx instead of context.Background(). Check:
// https://v3.wails.io/guide/services

// ----------------------------------------------------------------
// SourceService
// ----------------------------------------------------------------

// SourceService exposes browser source management and sync to the frontend.
type SourceService struct {
	app *app.App
	ctx context.Context
}

func NewSourceService(a *app.App) *SourceService {
	return &SourceService{app: a, ctx: context.Background()}
}

func (s *SourceService) DiscoverSources() ([]browser.DetectedSource, error) {
	found, errs := s.app.DiscoverSources()
	if len(errs) > 0 {
		// Return results with a soft warning — discovery errors are non-fatal
		// (one browser failing doesn't block others)
		return found, errs[0]
	}
	return found, nil
}

func (s *SourceService) DiscoverAndRegister() ([]dbgen.Source, error) {
	sources, errs := s.app.DiscoverAndRegister(s.ctx)
	if len(errs) > 0 {
		return sources, errs[0]
	}
	return sources, nil
}

func (s *SourceService) GetSources() ([]dbgen.Source, error) {
	return s.app.GetSources(s.ctx)
}

func (s *SourceService) SyncAll() []ingestion.SyncResult {
	return s.app.SyncAll(s.ctx)
}

func (s *SourceService) SyncSource(sourceID int64) (ingestion.SyncResult, error) {
	return s.app.SyncSource(s.ctx, sourceID)
}

func (s *SourceService) RemoveSource(sourceID int64) error {
	return s.app.RemoveSource(s.ctx, sourceID)
}

// ----------------------------------------------------------------
// DashboardService
// ----------------------------------------------------------------

// DashboardService exposes analytics queries to the frontend.
type DashboardService struct {
	app *app.App
	ctx context.Context
}

func NewDashboardService(a *app.App) *DashboardService {
	return &DashboardService{app: a, ctx: context.Background()}
}

func (s *DashboardService) GetDashboardStats(filter app.DashboardFilter) (dbgen.GetDashboardStatsRow, error) {
	return s.app.GetDashboardStats(s.ctx, filter)
}

func (s *DashboardService) GetTopDomains(filter app.DashboardFilter, limit int64) ([]dbgen.GetTopDomainsRow, error) {
	return s.app.GetTopDomains(s.ctx, filter, limit)
}

func (s *DashboardService) GetVisitTimeSeries(filter app.DashboardFilter) ([]dbgen.GetVisitTimeSeriesRow, error) {
	return s.app.GetVisitTimeSeries(s.ctx, filter)
}

func (s *DashboardService) GetActivityHeatmap(filter app.DashboardFilter) ([]idb.HeatmapCell, error) {
	// Timezone comes from the system — the frontend doesn't need to pass it
	return s.app.GetActivityHeatmap(s.ctx, filter, time.Local)
}

func (s *DashboardService) GetBrowserBreakdown(filter app.DashboardFilter) ([]dbgen.GetBrowserBreakdownRow, error) {
	return s.app.GetBrowserBreakdown(s.ctx, filter)
}

// ----------------------------------------------------------------
// SearchService
// ----------------------------------------------------------------

// SearchService exposes the search and query view to the frontend.
type SearchService struct {
	app *app.App
	ctx context.Context
}

func NewSearchService(a *app.App) *SearchService {
	return &SearchService{app: a, ctx: context.Background()}
}

func (s *SearchService) SearchVisits(params app.SearchParams) (app.SearchResults, error) {
	return s.app.SearchVisits(s.ctx, params)
}

// ----------------------------------------------------------------
// SettingsService
// ----------------------------------------------------------------

// SettingsService exposes settings read/write to the frontend.
type SettingsService struct {
	app *app.App
}

func NewSettingsService(a *app.App) *SettingsService {
	return &SettingsService{app: a}
}

func (s *SettingsService) GetSettings() config.Settings {
	return s.app.GetSettings()
}

func (s *SettingsService) UpdateSettings(settings config.Settings) error {
	return s.app.UpdateSettings(settings)
}