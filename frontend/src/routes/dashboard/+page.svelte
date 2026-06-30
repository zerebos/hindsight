<script lang="ts">
import {
    GetDashboardStats,
    GetTopDomains,
    GetVisitTimeSeries,
    GetActivityHeatmap,
    GetBrowserBreakdown,
} from '$hindsight/dashboardservice'
import { dashboardState } from '$lib/stores/dashboard.svelte'
import { appState } from '$lib/stores/app.svelte'
import { formatDuration, nullStr, formatHour, paletteColor, DAY_LABELS } from '$lib/utils'
import { hourTotals, weekdayTotals } from '$lib/analytics'
import TimeRangePicker from '$lib/components/TimeRangePicker.svelte'
import StatCard from '$lib/components/StatCard.svelte'
import Heatmap from '$lib/components/HeatMap.svelte'
import TimeSeries from '$lib/components/TimeSeries.svelte'
import Insights from '$lib/components/Insights.svelte'
import Donut from '$lib/components/Donut.svelte'
import DistributionBars from '$lib/components/DistributionBars.svelte'
import type { DashboardFilter } from '$happ/models'

async function loadDashboard(filter: DashboardFilter) {
    dashboardState.loading = true
    dashboardState.error = null

    try {
        const [stats, domains, series, heatmap, breakdown] = await Promise.all([
            GetDashboardStats(filter),
            GetTopDomains(filter, 15),
            GetVisitTimeSeries(filter),
            GetActivityHeatmap(filter),
            GetBrowserBreakdown(filter),
        ])

        dashboardState.stats            = stats
        dashboardState.topDomains       = domains
        dashboardState.timeSeries       = series
        dashboardState.heatmap          = heatmap
        dashboardState.browserBreakdown = breakdown
    } catch (err) {
        dashboardState.error = String(err)
    } finally {
        dashboardState.loading = false
    }
}

// Read filter synchronously before any await so $effect tracks it
$effect(() => {
    const filter = dashboardState.filter
    loadDashboard(filter)
})

// Reload when invalidated after sync
$effect(() => {
    if (dashboardState.stats === null && !dashboardState.loading && appState.initialized) {
        loadDashboard(dashboardState.filter)
    }
})

function onFilterChange(filter: DashboardFilter) {
    dashboardState.filter = filter
}

const totalBreakdownVisits = $derived(
    dashboardState.browserBreakdown.reduce((sum, b) => sum + b.TotalVisits, 0)
)

const browserSegments = $derived(
    dashboardState.browserBreakdown.map((b, i) => ({
        label: nullStr(b.Label, b.Browser),
        value: b.TotalVisits,
        color: paletteColor(i),
    }))
)

const hourLabels = Array.from({ length: 24 }, (_, h) => formatHour(h))
const hourValues = $derived(hourTotals(dashboardState.heatmap))
const weekdayValues = $derived(weekdayTotals(dashboardState.heatmap))

const statsLoading = $derived(dashboardState.loading && !dashboardState.stats)
</script>

<div class="dashboard">
    <header class="dashboard-header">
        <h1>Dashboard</h1>
        <TimeRangePicker
            filter={dashboardState.filter}
            onchange={onFilterChange}
        />
    </header>

    {#if dashboardState.error}
        <div class="page-error">{dashboardState.error}</div>
    {/if}

    <!-- Top-line KPIs -->
    <div class="stat-strip">
        <StatCard
            label="Total Visits"
            value={dashboardState.stats?.TotalVisits ?? null}
            loading={statsLoading}
        />
        <StatCard
            label="Unique Domains"
            value={dashboardState.stats?.UniqueDomains ?? null}
            loading={statsLoading}
        />
        <StatCard
            label="Active Days"
            value={dashboardState.stats?.ActiveDays ?? null}
            loading={statsLoading}
        />
        <StatCard
            label="Total Duration"
            value={formatDuration(dashboardState.stats?.TotalDurationMs ?? 0, {maxUnits: 2, short: true})}
            loading={statsLoading}
        />
        <StatCard
            label="Sources"
            value={appState.sources.length}
        />
    </div>

    <!-- Derived highlights -->
    <Insights
        timeSeries={dashboardState.timeSeries}
        heatmap={dashboardState.heatmap}
        topDomains={dashboardState.topDomains}
        stats={dashboardState.stats}
        loading={dashboardState.loading && dashboardState.heatmap.length === 0}
    />

    <!-- Visit history -->
    <section class="section">
        <h2 class="section-title">Visit History</h2>
        <TimeSeries
            data={dashboardState.timeSeries}
            loading={dashboardState.loading && dashboardState.timeSeries.length === 0}
            height={220}
        />
    </section>

    <!-- Domains + browsers -->
    <div class="main-grid">
        <section class="section domains-section">
            <h2 class="section-title">Top Domains</h2>
            {#if dashboardState.loading && dashboardState.topDomains.length === 0}
                {#each Array(8) as _, i (i)}
                    <div class="loading-placeholder" style="height: 28px; margin-bottom: 2px; border-radius: 3px;"></div>
                {/each}
            {:else}
                <table class="data-table">
                    <thead>
                        <tr>
                            <th>#</th>
                            <th>Domain</th>
                            <th style="text-align: right;">Share</th>
                            <th style="text-align: right; width: 80px;">Visits</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each dashboardState.topDomains as domain, i (domain.Host)}
                            {@const share = dashboardState.stats && dashboardState.stats.TotalVisits > 0
                                ? (domain.TotalVisits / dashboardState.stats.TotalVisits * 100)
                                : 0}
                            <tr>
                                <td class="rank">{i + 1}</td>
                                <td class="domain-cell">
                                    <span class="domain-name">{domain.Host}</span>
                                </td>
                                <td class="share-cell">
                                    <div class="share-bar-wrap">
                                        <div class="share-bar-track">
                                            <div class="share-bar" style="width: {Math.max(2, share)}%"></div>
                                        </div>
                                        <span class="share-pct">{share.toFixed(1)}%</span>
                                    </div>
                                </td>
                                <td class="visits-cell">{domain.TotalVisits.toLocaleString()}</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            {/if}
        </section>

        <section class="section breakdown-section">
            <h2 class="section-title">Browsers</h2>
            {#if dashboardState.loading && dashboardState.browserBreakdown.length === 0}
                <div class="loading-placeholder" style="height: 160px; border-radius: 50%; width: 160px; margin: 0 auto;"></div>
            {:else}
                <Donut
                    segments={browserSegments}
                    centerLabel="visits"
                    size={160}
                />
            {/if}
        </section>
    </div>

    <!-- Heatmap -->
    <section class="section">
        <h2 class="section-title">Activity Heatmap</h2>
        <p class="section-subtitle">When you browse · day of week × hour of day · your current timezone</p>
        <Heatmap
            data={dashboardState.heatmap}
            loading={dashboardState.loading && dashboardState.heatmap.length === 0}
        />
    </section>

    <!-- Rhythm: hour + weekday distributions -->
    <div class="main-grid even">
        <section class="section">
            <h2 class="section-title">Busiest Hours</h2>
            <DistributionBars
                values={hourValues}
                labels={hourLabels}
                tickEvery={3}
                loading={dashboardState.loading && dashboardState.heatmap.length === 0}
                height={130}
            />
        </section>

        <section class="section">
            <h2 class="section-title">Busiest Days</h2>
            <DistributionBars
                values={weekdayValues}
                labels={[...DAY_LABELS]}
                loading={dashboardState.loading && dashboardState.heatmap.length === 0}
                height={130}
            />
        </section>
    </div>
</div>

<style>
    .dashboard {
        display: flex;
        flex-direction: column;
        gap: var(--space-4);
        padding: var(--space-4) var(--space-6);
        height: 100%;
        overflow-y: auto;
    }

    .dashboard-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        flex-shrink: 0;
    }

    .dashboard-header h1 {
        font-size: var(--font-size-lg);
        font-weight: 600;
    }

    .page-error {
        padding: var(--space-3);
        background: var(--error-subtle);
        border: 1px solid var(--error);
        border-radius: var(--radius-sm);
        color: var(--error);
        font-size: var(--font-size-sm);
    }

    .stat-strip {
        display: flex;
        gap: var(--space-3);
        flex-shrink: 0;
        flex-wrap: wrap;
    }

    .main-grid {
        display: grid;
        grid-template-columns: 1fr 300px;
        gap: var(--space-4);
        align-items: start;
    }

    .main-grid.even {
        grid-template-columns: 1fr 1fr;
    }

    .section {
        background: var(--surface);
        border: 1px solid var(--border);
        border-radius: var(--radius);
        padding: var(--space-3) var(--space-4);
    }

    .breakdown-section {
        display: flex;
        flex-direction: column;
    }

    .section-title {
        font-size: var(--font-size-sm);
        font-weight: 600;
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
        margin-bottom: var(--space-3);
    }

    .section-subtitle {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
        margin-top: -10px;
        margin-bottom: var(--space-3);
    }

    .rank {
        color: var(--text-faint);
        font-size: var(--font-size-xs);
        width: 24px;
        font-variant-numeric: tabular-nums;
    }

    .domain-cell { max-width: 240px; }

    .domain-name {
        font-family: var(--font-mono);
        font-size: var(--font-size-xs);
        color: var(--text);
    }

    .visits-cell {
        text-align: right;
        font-variant-numeric: tabular-nums;
        font-size: var(--font-size-sm);
        color: var(--text-muted);
    }

    .share-cell { text-align: right; width: 130px; }

    .share-bar-wrap {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: var(--space-2);
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }

    .share-bar-track {
        flex: 1;
        max-width: 64px;
        height: 4px;
        background: var(--surface-2);
        border-radius: 2px;
        overflow: hidden;
    }

    .share-bar {
        height: 100%;
        background: var(--accent);
        border-radius: 2px;
        opacity: 0.6;
        min-width: 1px;
    }

    .share-pct {
        font-variant-numeric: tabular-nums;
        width: 3.2ch;
        text-align: right;
    }
</style>
