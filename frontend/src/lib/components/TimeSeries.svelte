<script lang="ts">
import type { GetVisitTimeSeriesRow } from '$dbgen/models'
import { formatShortDate, formatNumber } from '$lib/utils'

interface Props {
    data: GetVisitTimeSeriesRow[]
    loading?: boolean
    height?: number
}

let { data, loading = false, height = 120 }: Props = $props()

const LABEL_H  = 20
const LABEL_W  = 40
const BAR_GAP  = 1

const chartHeight = $derived(height - LABEL_H)

const maxVisits = $derived(
    data.length === 0 ? 1 : Math.max(...data.map(d => d.TotalVisits))
)

// Only show every Nth date label to avoid crowding
const labelInterval = $derived(Math.ceil(data.length / 8))

function barHeight(visits: number): number {
    return Math.max(1, (visits / maxVisits) * chartHeight)
}

// Tooltip state
let tooltip = $state<{ x: number; y: number; date: string; visits: number } | null>(null)

function showTooltip(event: MouseEvent, row: GetVisitTimeSeriesRow) {
    const rect = (event.target as SVGElement).getBoundingClientRect()
    tooltip = {
        x: rect.x + rect.width / 2,
        y: rect.y - 8,
        date: formatShortDate(row.Day),
        visits: row.TotalVisits,
    }
}

function hideTooltip() {
    tooltip = null
}
</script>

<div class="timeseries-wrapper" style="height: {height}px;">
    {#if loading}
        <div class="loading-placeholder" style="height: 100%; border-radius: var(--radius-sm);"></div>
    {:else if data.length === 0}
        <div class="empty">No data for this period</div>
    {:else}
        <svg
            width="100%"
            height={height}
            viewBox="0 0 {data.length * (BAR_GAP + 4) + LABEL_W} {height}"
            preserveAspectRatio="none"
            class="timeseries-svg"
            aria-label="Visit time series"
        >
            {#each data as row, i}
                {@const bh = barHeight(row.TotalVisits)}
                {@const bx = LABEL_W + i * (4 + BAR_GAP)}
                {@const by = chartHeight - bh}

                <rect
                    x={bx}
                    y={by}
                    width="4"
                    height={bh}
                    rx="1"
                    class="bar"
                    onmouseenter={(e) => showTooltip(e, row)}
                    onmouseleave={hideTooltip}
                    aria-label="{formatShortDate(row.Day)}: {row.TotalVisits} visits"
                />

                {#if i % labelInterval === 0}
                    <text
                        x={bx + 2}
                        y={height - 2}
                        class="axis-label"
                        text-anchor="middle"
                    >{formatShortDate(row.Day).slice(5)}</text>
                {/if}
            {/each}

            <!-- Y axis max label -->
            <text x={LABEL_W - 4} y={8} class="axis-label" text-anchor="end">
                {formatNumber(maxVisits)}
            </text>
        </svg>

        {#if tooltip}
            <div
                class="tooltip"
                style="left: {tooltip.x}px; top: {tooltip.y}px;"
            >
                <strong>{tooltip.date}</strong>
                <span>{formatNumber(tooltip.visits)} visits</span>
            </div>
        {/if}
    {/if}
</div>

<style>
    .timeseries-wrapper {
        position: relative;
        width: 100%;
    }

    .timeseries-svg {
        display: block;
        width: 100%;
        overflow: visible;
    }

    .bar {
        fill: var(--accent);
        opacity: 0.7;
        cursor: default;
        transition: opacity 0.1s;
    }

    .bar:hover {
        opacity: 1;
    }

    .axis-label {
        fill: var(--text-faint);
        font-size: 9px;
        font-family: var(--font-sans);
    }

    .empty {
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--text-faint);
        font-size: var(--font-size-sm);
    }

    .tooltip {
        position: fixed;
        transform: translate(-50%, -100%);
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: var(--radius-sm);
        padding: 4px 8px;
        font-size: var(--font-size-xs);
        color: var(--text);
        pointer-events: none;
        white-space: nowrap;
        z-index: 100;
        display: flex;
        flex-direction: column;
        gap: 1px;
    }
</style>