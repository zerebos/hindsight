import type {DashboardFilter, DomainInsights, TimeSpent} from '$hindsight/internal/app/models';
import type {
    GetDashboardStatsRow,
    GetTopDomainsRow,
    GetVisitTimeSeriesRow,
    GetBrowserBreakdownRow,
} from '$hindsight/internal/db/generated/models';
import type {HeatmapCell, TrackingStats} from '$hindsight/internal/db/models';

export const dashboardState = $state({
    filter: {StartTime: 0, EndTime: 0} as DashboardFilter,
    stats: null as GetDashboardStatsRow | null,
    topDomains: [] as GetTopDomainsRow[],
    timeSeries: [] as GetVisitTimeSeriesRow[],
    heatmap: [] as HeatmapCell[],
    browserBreakdown: [] as GetBrowserBreakdownRow[],
    tracking: null as TrackingStats | null,
    domainInsights: null as DomainInsights | null,
    timeSpent: null as TimeSpent | null,
    // Previous equal-length window, for period-over-period trends. Null when
    // the current filter is unbounded ("All"), where a comparison is undefined.
    prevStats: null as GetDashboardStatsRow | null,
    prevTopDomains: [] as GetTopDomainsRow[],
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
    dashboardState.tracking = null;
    dashboardState.domainInsights = null;
    dashboardState.timeSpent = null;
    dashboardState.prevStats = null;
    dashboardState.prevTopDomains = [];
}
