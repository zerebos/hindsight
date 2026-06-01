import type {DashboardFilter} from '$hindsight/internal/app/models';
import type {
    GetDashboardStatsRow,
    GetTopDomainsRow,
    GetVisitTimeSeriesRow,
    GetBrowserBreakdownRow,
} from '$hindsight/internal/db/generated/models';
import type {HeatmapCell} from '$hindsight/internal/db/models';

export const dashboardState = $state({
    filter: {StartTime: 0, EndTime: 0} as DashboardFilter,
    stats: null as GetDashboardStatsRow | null,
    topDomains: [] as GetTopDomainsRow[],
    timeSeries: [] as GetVisitTimeSeriesRow[],
    heatmap: [] as HeatmapCell[],
    browserBreakdown: [] as GetBrowserBreakdownRow[],
    loading: false,
    error: null as string | null,
});

/** Invalidate all cached data — triggers a reload on next dashboard visit */
export function invalidateDashboard() {
    dashboardState.stats = null;
    dashboardState.topDomains = [];
    dashboardState.timeSeries = [];
    dashboardState.heatmap = [];
    dashboardState.browserBreakdown = [];
}