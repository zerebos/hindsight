<script lang="ts">
import type { TrackingStats } from '$db/models'
import { formatNumber, formatPercent } from '$lib/utils'

interface Props {
    tracking: TrackingStats | null
    totalVisits: number
    loading?: boolean
}

let { tracking, totalVisits, loading = false }: Props = $props()

const share = $derived(
    tracking && totalVisits > 0 ? tracking.TrackedVisits / totalVisits : 0,
)
const maxParam = $derived(
    tracking && tracking.TopParams.length > 0 ? tracking.TopParams[0].Count : 1,
)
</script>

<div class="tracking">
    {#if loading}
        <div class="loading-placeholder" style="height: 140px;"></div>
    {:else if !tracking || tracking.TrackedVisits === 0}
        <div class="empty">No tracking parameters found in this period.</div>
    {:else}
        <div class="headline">
            <div class="pct-block">
                <span class="pct">{formatPercent(share)}</span>
                <span class="pct-label">of visits carried trackers</span>
            </div>
            <span class="count">{formatNumber(tracking.TrackedVisits)} tracked</span>
        </div>

        <div class="param-block">
            <span class="block-label">Top parameters</span>
            <div class="params">
                {#each tracking.TopParams as p (p.Name)}
                    <div class="param">
                        <span class="param-name" title={p.Name}>{p.Name}</span>
                        <div class="param-track">
                            <div class="param-bar" style="width: {(p.Count / maxParam) * 100}%"></div>
                        </div>
                        <span class="param-count">{formatNumber(p.Count)}</span>
                    </div>
                {/each}
            </div>
        </div>

        {#if tracking.TopDomains.length > 0}
            <div class="domain-block">
                <span class="block-label">Most-tracked domains</span>
                <div class="domain-chips">
                    {#each tracking.TopDomains.slice(0, 5) as d (d.Host)}
                        <span class="chip" title="{d.Host}: {formatNumber(d.Count)} tracked visits">
                            {d.Host}<span class="chip-count">{formatNumber(d.Count)}</span>
                        </span>
                    {/each}
                </div>
            </div>
        {/if}
    {/if}
</div>

<style>
    .tracking {
        display: flex;
        flex-direction: column;
        gap: var(--space-3);
    }

    .headline {
        display: flex;
        align-items: baseline;
        justify-content: space-between;
        gap: var(--space-3);
    }

    .pct-block {
        display: flex;
        align-items: baseline;
        gap: var(--space-2);
    }

    .pct {
        font-size: 1.6rem;
        font-weight: 600;
        color: var(--warning);
        font-variant-numeric: tabular-nums;
    }

    .pct-label {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }

    .count {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
        flex-shrink: 0;
    }

    .block-label {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .param-block, .domain-block {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
    }

    .params {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .param {
        display: grid;
        grid-template-columns: 90px 1fr auto;
        align-items: center;
        gap: var(--space-2);
    }

    .param-name {
        font-family: var(--font-mono);
        font-size: var(--font-size-xs);
        color: var(--text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .param-track {
        height: 6px;
        background: var(--surface-2);
        border-radius: 3px;
        overflow: hidden;
    }

    .param-bar {
        height: 100%;
        background: var(--warning);
        opacity: 0.7;
        border-radius: 3px;
        min-width: 2px;
    }

    .param-count {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
    }

    .domain-chips {
        display: flex;
        flex-wrap: wrap;
        gap: var(--space-2);
    }

    .chip {
        display: inline-flex;
        align-items: center;
        gap: var(--space-2);
        padding: 3px 8px;
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: 999px;
        font-family: var(--font-mono);
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        max-width: 100%;
    }

    .chip-count {
        color: var(--text-faint);
        font-variant-numeric: tabular-nums;
    }

    .empty {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 120px;
        color: var(--text-faint);
        font-size: var(--font-size-sm);
        text-align: center;
    }
</style>
