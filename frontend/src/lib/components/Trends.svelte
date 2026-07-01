<script lang="ts">
import type { GetDashboardStatsRow, GetTopDomainsRow } from '$dbgen/models'
import { pctChange, rankMovers } from '$lib/analytics'
import { formatNumber, formatSigned } from '$lib/utils'

interface Props {
    stats: GetDashboardStatsRow | null
    prevStats: GetDashboardStatsRow | null
    topDomains: GetTopDomainsRow[]
    prevTopDomains: GetTopDomainsRow[]
    rangeLabel: string
    loading?: boolean
}

let { stats, prevStats, topDomains, prevTopDomains, rangeLabel, loading = false }: Props = $props()

interface Metric {
    label: string
    value: number
    change: number | null
}

const metrics = $derived.by<Metric[]>(() => {
    if (!stats || !prevStats) return []
    return [
        { label: 'Visits', value: stats.TotalVisits, change: pctChange(stats.TotalVisits, prevStats.TotalVisits) },
        { label: 'Domains', value: stats.UniqueDomains, change: pctChange(stats.UniqueDomains, prevStats.UniqueDomains) },
        { label: 'Active days', value: stats.ActiveDays, change: pctChange(stats.ActiveDays, prevStats.ActiveDays) },
    ]
})

const movers = $derived(
    prevTopDomains.length > 0 ? rankMovers(topDomains, prevTopDomains, 4) : [],
)

function tone(change: number | null): string {
    if (change === null || change === 0) return 'flat'
    return change > 0 ? 'up' : 'down'
}
</script>

{#if loading}
    <div class="trends loading-placeholder" style="height: 84px;"></div>
{:else if stats && prevStats}
    <div class="trends">
        <div class="metrics">
            <span class="trends-caption">{rangeLabel}</span>
            <div class="metric-row">
                {#each metrics as m (m.label)}
                    <div class="metric">
                        <span class="metric-label">{m.label}</span>
                        <span class="metric-value">{formatNumber(m.value)}</span>
                        <span class="metric-change {tone(m.change)}">
                            {#if m.change === null}—{:else}{formatSigned(m.change)}{/if}
                        </span>
                    </div>
                {/each}
            </div>
        </div>

        {#if movers.length > 0}
            <div class="movers">
                <span class="trends-caption">Rank movers</span>
                <div class="mover-list">
                    {#each movers as mv (mv.host)}
                        <div class="mover">
                            <span class="mover-arrow up">▲</span>
                            <span class="mover-host" title={mv.host}>{mv.host}</span>
                            <span class="mover-delta">{mv.isNew ? 'new' : `+${mv.delta}`}</span>
                        </div>
                    {/each}
                </div>
            </div>
        {/if}
    </div>
{/if}

<style>
    .trends {
        display: flex;
        gap: var(--space-6);
        align-items: stretch;
        padding: var(--space-3) var(--space-4);
        background: var(--surface);
        border: 1px solid var(--border);
        border-radius: var(--radius);
        flex-wrap: wrap;
    }

    .metrics {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
        flex: 1;
        min-width: 260px;
    }

    .trends-caption {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
        font-weight: 500;
    }

    .metric-row {
        display: flex;
        gap: var(--space-6);
        flex-wrap: wrap;
    }

    .metric {
        display: grid;
        grid-template-columns: auto auto;
        align-items: baseline;
        gap: 2px var(--space-2);
    }

    .metric-label {
        grid-column: 1 / -1;
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }

    .metric-value {
        font-size: 1.35rem;
        font-weight: 600;
        font-variant-numeric: tabular-nums;
        color: var(--text);
    }

    .metric-change {
        font-size: var(--font-size-sm);
        font-weight: 600;
        font-variant-numeric: tabular-nums;
    }

    .metric-change.up { color: var(--success); }
    .metric-change.down { color: var(--error); }
    .metric-change.flat { color: var(--text-faint); }

    .movers {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
        min-width: 200px;
    }

    .mover-list {
        display: flex;
        flex-direction: column;
        gap: 3px;
    }

    .mover {
        display: flex;
        align-items: center;
        gap: var(--space-2);
        font-size: var(--font-size-sm);
    }

    .mover-arrow.up { color: var(--success); font-size: 9px; }

    .mover-host {
        font-family: var(--font-mono);
        font-size: var(--font-size-xs);
        color: var(--text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 160px;
        flex: 1;
    }

    .mover-delta {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
        flex-shrink: 0;
    }
</style>
