<script lang="ts">
import { onMount } from 'svelte'
import {
    GetDashboardStats,
    GetTopDomains,
    GetVisitTimeSeries,
    GetActivityHeatmap,
    GetBrowserBreakdown,
} from '$hindsight/dashboardservice'
import { dashboardState } from '$lib/stores/dashboard.svelte'
import { appState } from '$lib/stores/app.svelte'
import { formatDuration, nullStr } from '$lib/utils'
import TimeRangePicker from '$lib/components/TimeRangePicker.svelte'
import StatCard from '$lib/components/StatCard.svelte'
import Heatmap from '$lib/components/HeatMap.svelte'
import TimeSeries from '$lib/components/TimeSeries.svelte'
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

    <div class="stat-strip">
        <StatCard
            label="Total Visits"
            value={dashboardState.stats?.TotalVisits ?? null}
            loading={dashboardState.loading && !dashboardState.stats}
        />
        <StatCard
            label="Unique Domains"
            value={dashboardState.stats?.UniqueDomains ?? null}
            loading={dashboardState.loading && !dashboardState.stats}
        />
        <StatCard
            label="Active Days"
            value={dashboardState.stats?.ActiveDays ?? null}
            loading={dashboardState.loading && !dashboardState.stats}
        />
        <StatCard
            label="Total Duration"
            value={formatDuration(dashboardState.stats?.TotalDurationMs ?? 0, {maxUnits: 2, short: true})}
            loading={dashboardState.loading && !dashboardState.stats}
        />
        <StatCard
            label="Sources"
            value={appState.sources.length}
        />
    </div>

    <div class="main-grid">
        <section class="section domains-section">
            <h2 class="section-title">Top Domains</h2>
            {#if dashboardState.loading && dashboardState.topDomains.length === 0}
                {#each Array(8) as _}
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
                        {#each dashboardState.topDomains as domain, i}
                            {@const share = dashboardState.stats
                                ? (domain.TotalVisits / dashboardState.stats.TotalVisits * 100).toFixed(1)
                                : '—'}
                            <tr>
                                <td class="rank">{i + 1}</td>
                                <td class="domain-cell">
                                    <span class="domain-name">{domain.Host}</span>
                                </td>
                                <td class="share-cell">
                                    <div class="share-bar-wrap">
                                        <div class="share-bar" style="width: calc(5 * {share}px)"></div>
                                        <span style:margin-left={share.length === 3 ? '1ch' : '0'}>{share}%</span>
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
                {#each Array(4) as _}
                    <div class="loading-placeholder" style="height: 28px; margin-bottom: 2px; border-radius: 3px;"></div>
                {/each}
            {:else}
                <div class="breakdown-list">
                    {#each dashboardState.browserBreakdown as browser}
                        {@const pct = totalBreakdownVisits > 0
                            ? (browser.TotalVisits / totalBreakdownVisits * 100)
                            : 0}
                        <div class="breakdown-row">
                            <div class="breakdown-meta">
                                <span class="breakdown-label">{nullStr(browser.Label, browser.Browser)}</span>
                                <span class="breakdown-visits text-muted">{browser.TotalVisits.toLocaleString()}</span>
                            </div>
                            <div class="breakdown-bar-track">
                                <div class="breakdown-bar" style="width: {pct.toFixed(1)}%"></div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </section>
    </div>

    <section class="section chart-section">
        <h2 class="section-title">Visit History</h2>
        <TimeSeries
            data={dashboardState.timeSeries}
            loading={dashboardState.loading && dashboardState.timeSeries.length === 0}
            height={140}
        />
    </section>

    <section class="section">
        <h2 class="section-title">Activity by Hour</h2>
        <p class="section-subtitle">Your current timezone · day of week × hour of day</p>
        <Heatmap
            data={dashboardState.heatmap}
            loading={dashboardState.loading && dashboardState.heatmap.length === 0}
        />
    </section>
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
    }

    .main-grid {
        display: grid;
        grid-template-columns: 1fr 280px;
        gap: var(--space-4);
        align-items: start;
    }

    .section {
        background: var(--surface);
        border: 1px solid var(--border);
        border-radius: var(--radius);
        padding: var(--space-3) var(--space-4);
    }

    .section.chart-section {
        /* overflow-x: auto; */
        overflow: hidden;
        min-height: 200px;
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

    .share-cell { text-align: right; }

    .share-bar-wrap {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: var(--space-2);
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }

    .share-bar {
        height: 3px;
        background: var(--accent);
        border-radius: 2px;
        opacity: 0.5;
        min-width: 1px;
    }

    .breakdown-list {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
    }

    .breakdown-row {
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .breakdown-meta {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
    }

    .breakdown-label {
        font-size: var(--font-size-sm);
        color: var(--text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 160px;
    }

    .breakdown-visits {
        font-size: var(--font-size-xs);
        font-variant-numeric: tabular-nums;
        flex-shrink: 0;
    }

    .breakdown-bar-track {
        height: 3px;
        background: var(--surface-2);
        border-radius: 2px;
        overflow: hidden;
    }

    .breakdown-bar {
        height: 100%;
        background: var(--accent);
        border-radius: 2px;
        opacity: 0.6;
    }
</style>