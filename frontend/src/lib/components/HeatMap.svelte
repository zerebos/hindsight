<script lang="ts">
import type { HeatmapCell } from '$db/models'
import { scaleQuantize } from 'd3-scale'
import { DAY_LABELS, DAY_LABELS_FULL, formatHour, formatNumber } from '$lib/utils'
import { weekdayTotals } from '$lib/analytics'

interface Props {
    data: HeatmapCell[]
    loading?: boolean
}

let { data, loading = false }: Props = $props()

const GAP = 3
const LABEL_W = 34   // left gutter for day names
const LABEL_H = 18   // top gutter for hour labels
const TOTAL_W = 46   // right gutter for per-day total bars
const CELL_MIN = 13
const CELL_MAX = 26

// Opacity steps for the 5 intensity buckets (GitHub-style ramp on one hue).
const LEVELS = [0.16, 0.38, 0.6, 0.8, 1]

let width = $state(720)

const cellSize = $derived(
    Math.max(
        CELL_MIN,
        Math.min(CELL_MAX, Math.floor((width - LABEL_W - TOTAL_W - 23 * GAP) / 24)),
    ),
)

const gridW = $derived(24 * cellSize + 23 * GAP)
const gridH = $derived(7 * cellSize + 6 * GAP)
const svgW = $derived(LABEL_W + gridW + TOTAL_W)
const svgH = $derived(LABEL_H + gridH)

const lookup = $derived.by(() => {
    const map = new Map<string, number>()
    for (const c of data) map.set(`${c.Day}-${c.Hour}`, c.TotalVisits)
    return map
})

const maxVisits = $derived(data.length === 0 ? 0 : Math.max(...data.map(c => c.TotalVisits)))

// Quantize visit counts into 5 buckets for a readable, stepped color scale.
const bucket = $derived(scaleQuantize<number>().domain([1, Math.max(1, maxVisits)]).range([0, 1, 2, 3, 4]))

const dayTotals = $derived(weekdayTotals(data))
const maxDayTotal = $derived(Math.max(1, ...dayTotals))

// Peak cell, highlighted with a ring.
const peak = $derived.by(() => {
    let best: HeatmapCell | null = null
    for (const c of data) if (!best || c.TotalVisits > best.TotalVisits) best = c
    return best
})

function cellValue(day: number, hour: number): number {
    return lookup.get(`${day}-${hour}`) ?? 0
}

function cellOpacity(visits: number): number {
    if (visits <= 0) return 0
    return LEVELS[bucket(visits)]
}

function cellX(hour: number): number {
    return LABEL_W + hour * (cellSize + GAP)
}

function cellY(day: number): number {
    return LABEL_H + day * (cellSize + GAP)
}

// ---- Tooltip ----
let tooltip = $state<{ x: number; y: number; day: number; hour: number; visits: number } | null>(null)

function showTooltip(event: MouseEvent, day: number, hour: number, visits: number) {
    const r = (event.target as SVGElement).getBoundingClientRect()
    tooltip = { x: r.x + r.width / 2, y: r.y - 8, day, hour, visits }
}

function hideTooltip() {
    tooltip = null
}
</script>

<div class="heatmap-wrapper" bind:clientWidth={width}>
    {#if loading}
        <div class="loading-placeholder" style="height: 168px; border-radius: var(--radius-sm);"></div>
    {:else if data.length === 0}
        <div class="empty">No activity in this period</div>
    {:else}
        <svg width="100%" viewBox="0 0 {svgW} {svgH}" class="heatmap-svg" aria-label="Activity heatmap by day of week and hour of day">
            <!-- Day labels -->
            {#each DAY_LABELS as day, i (i)}
                <text x={LABEL_W - 8} y={cellY(i) + cellSize / 2} class="axis-label" text-anchor="end" dominant-baseline="middle">{day}</text>
            {/each}

            <!-- Hour labels every 3 hours -->
            {#each Array.from({ length: 9 }, (_, i) => i * 3) as hour (hour)}
                {#if hour <= 24}
                    <text x={cellX(hour)} y={LABEL_H - 6} class="axis-label" text-anchor="middle">{hour === 24 ? '' : formatHour(hour)}</text>
                {/if}
            {/each}

            <!-- Cells -->
            {#each Array.from({ length: 7 }, (_, d) => d) as day (day)}
                {#each Array.from({ length: 24 }, (_, h) => h) as hour (hour)}
                    {@const visits = cellValue(day, hour)}
                    {@const isPeak = peak !== null && peak.Day === day && peak.Hour === hour}
                    <rect
                        x={cellX(hour)}
                        y={cellY(day)}
                        width={cellSize}
                        height={cellSize}
                        rx="2.5"
                        class="cell"
                        class:has-visits={visits > 0}
                        class:peak={isPeak}
                        opacity={visits > 0 ? cellOpacity(visits) : 1}
                        onmouseenter={(e) => showTooltip(e, day, hour, visits)}
                        onmouseleave={hideTooltip}
                        tabindex="-1"
                        role="gridcell"
                        aria-label="{DAY_LABELS_FULL[day]} {formatHour(hour)}: {visits} visits"
                    />
                {/each}

                <!-- Per-day total bar (right margin) -->
                {@const total = dayTotals[day]}
                {@const barW = (total / maxDayTotal) * (TOTAL_W - 8)}
                <rect
                    x={LABEL_W + gridW + 6}
                    y={cellY(day) + cellSize / 2 - 3}
                    width={Math.max(total > 0 ? 2 : 0, barW)}
                    height="6"
                    rx="2"
                    class="total-bar"
                />
            {/each}
        </svg>

        <div class="legend">
            <span class="legend-label">Less</span>
            <div class="legend-swatch empty-swatch"></div>
            {#each LEVELS as op (op)}
                <div class="legend-swatch" style="opacity: {op};"></div>
            {/each}
            <span class="legend-label">More</span>
            <span class="legend-spacer"></span>
            <span class="legend-label peak-key">Peak: {peak ? `${DAY_LABELS[peak.Day]} ${formatHour(peak.Hour)}` : '—'}</span>
        </div>

        {#if tooltip}
            <div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">
                <strong>{DAY_LABELS_FULL[tooltip.day]} · {formatHour(tooltip.hour)}</strong>
                <span>{formatNumber(tooltip.visits)} visits</span>
            </div>
        {/if}
    {/if}
</div>

<style>
    .heatmap-wrapper {
        position: relative;
        width: 100%;
    }

    .heatmap-svg {
        display: block;
        overflow: visible;
    }

    .axis-label {
        fill: var(--text-faint);
        font-size: 10px;
        font-family: var(--font-sans);
    }

    .cell {
        fill: var(--surface-2);
        cursor: default;
        transition: opacity 0.1s;
    }

    .cell.has-visits {
        fill: var(--accent);
    }

    .cell:hover {
        stroke: var(--text);
        stroke-width: 1;
    }

    .cell.peak {
        stroke: var(--accent-hover);
        stroke-width: 1.5;
    }

    .total-bar {
        fill: var(--accent);
        opacity: 0.45;
    }

    .legend {
        display: flex;
        align-items: center;
        gap: 4px;
        margin-top: var(--space-3);
        font-size: var(--font-size-xs);
    }

    .legend-swatch {
        width: 13px;
        height: 13px;
        border-radius: 2.5px;
        background: var(--accent);
    }

    .legend-swatch.empty-swatch {
        background: var(--surface-2);
        opacity: 1;
    }

    .legend-label {
        color: var(--text-faint);
    }

    .legend-spacer {
        flex: 1;
    }

    .peak-key {
        color: var(--text-muted);
    }

    .empty {
        height: 120px;
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
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
    }
</style>
