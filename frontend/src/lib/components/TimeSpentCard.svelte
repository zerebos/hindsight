<script lang="ts">
import type { TimeSpent } from '$happ/models'
import { formatDuration, formatPercent, formatNumber } from '$lib/utils'

interface Props {
    timeSpent: TimeSpent | null
    loading?: boolean
}

let { timeSpent, loading = false }: Props = $props()

const coverage = $derived(
    timeSpent && timeSpent.TotalVisits > 0 ? timeSpent.VisitsWithDuration / timeSpent.TotalVisits : 0,
)
const avgMs = $derived(
    timeSpent && timeSpent.VisitsWithDuration > 0
        ? timeSpent.TotalDurationMs / timeSpent.VisitsWithDuration
        : 0,
)
const maxDomainMs = $derived(
    timeSpent && timeSpent.TopDomains.length > 0 ? timeSpent.TopDomains[0].TotalDurationMs : 1,
)
</script>

<div class="time-spent">
    {#if loading}
        <div class="loading-placeholder" style="height: 140px;"></div>
    {:else if !timeSpent || timeSpent.VisitsWithDuration === 0}
        <div class="empty">
            No visit-duration data in this period.<br />
            <span class="hint">Not every browser reports how long pages are open.</span>
        </div>
    {:else}
        <div class="headline">
            <div class="total-block">
                <span class="total">{formatDuration(timeSpent.TotalDurationMs, { maxUnits: 2, short: true })}</span>
                <span class="total-label">total time</span>
            </div>
            <div class="meta">
                <span class="meta-item">~{formatDuration(avgMs, { maxUnits: 1, short: true })} avg / visit</span>
                <span class="meta-item">{formatPercent(coverage)} of visits timed</span>
            </div>
        </div>

        <div class="domain-block">
            <span class="block-label">Where time went</span>
            <div class="domains">
                {#each timeSpent.TopDomains as d (d.Host)}
                    <div class="domain-row">
                        <span class="domain-name" title={d.Host}>{d.Host}</span>
                        <div class="domain-track">
                            <div class="domain-bar" style="width: {(d.TotalDurationMs / maxDomainMs) * 100}%"></div>
                        </div>
                        <span class="domain-time">{formatDuration(d.TotalDurationMs, { maxUnits: 1, short: true })}</span>
                    </div>
                {/each}
            </div>
        </div>
    {/if}
</div>

<style>
    .time-spent {
        display: flex;
        flex-direction: column;
        gap: var(--space-3);
    }

    .headline {
        display: flex;
        align-items: baseline;
        justify-content: space-between;
        gap: var(--space-3);
        flex-wrap: wrap;
    }

    .total-block {
        display: flex;
        align-items: baseline;
        gap: var(--space-2);
    }

    .total {
        font-size: 1.6rem;
        font-weight: 600;
        color: var(--accent);
        font-variant-numeric: tabular-nums;
    }

    .total-label {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }

    .meta {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        gap: 1px;
    }

    .meta-item {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
    }

    .block-label {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .domain-block {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
    }

    .domains {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .domain-row {
        display: grid;
        grid-template-columns: 150px 1fr auto;
        align-items: center;
        gap: var(--space-2);
    }

    .domain-name {
        font-family: var(--font-mono);
        font-size: var(--font-size-xs);
        color: var(--text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .domain-track {
        height: 6px;
        background: var(--surface-2);
        border-radius: 3px;
        overflow: hidden;
    }

    .domain-bar {
        height: 100%;
        background: var(--accent);
        opacity: 0.65;
        border-radius: 3px;
        min-width: 2px;
    }

    .domain-time {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
    }

    .empty {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 120px;
        color: var(--text-faint);
        font-size: var(--font-size-sm);
        text-align: center;
        gap: 4px;
    }

    .hint {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }
</style>
