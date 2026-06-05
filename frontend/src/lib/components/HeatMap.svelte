<script lang="ts">
import type { HeatmapCell } from '$db/models'
import { DAY_LABELS, formatHour } from '$lib/utils'

interface Props {
    data: HeatmapCell[]
    loading?: boolean
}

let { data, loading = false }: Props = $props()

const CELL_SIZE = 14
const CELL_GAP  = 3
const LABEL_W   = 30  // space for hour labels on left
const LABEL_H   = 20  // space for day labels on top

// Build a 7×24 grid — rows=days, cols=hours
const gridData = $derived.by(() => {
    const map = new Map<string, number>()
    for (const cell of data) {
        map.set(`${cell.Day}-${cell.Hour}`, cell.TotalVisits)
    }
    return map
})

const maxVisits = $derived.by(() => {
    if (data.length === 0) return 1
    return Math.max(...data.map(c => c.TotalVisits))
})

function cellValue(day: number, hour: number): number {
    return gridData.get(`${day}-${hour}`) ?? 0
}

function cellOpacity(visits: number): number {
    if (visits === 0) return 0
    // Minimum visible opacity 0.15, max 1.0
    return 0.15 + (visits / maxVisits) * 0.85
}

function cellX(hour: number): number {
    return LABEL_W + hour * (CELL_SIZE + CELL_GAP)
}

function cellY(day: number): number {
    return LABEL_H + day * (CELL_SIZE + CELL_GAP)
}

const svgWidth  = $derived(LABEL_W + 24 * (CELL_SIZE + CELL_GAP))
const svgHeight = $derived(LABEL_H + 7  * (CELL_SIZE + CELL_GAP))

// Tooltip state
let tooltip = $state<{ x: number; y: number; day: number; hour: number; visits: number } | null>(null)

function showTooltip(event: MouseEvent, day: number, hour: number, visits: number) {
    tooltip = {
        x: (event.target as SVGElement).getBoundingClientRect().x + CELL_SIZE / 2,
        y: (event.target as SVGElement).getBoundingClientRect().y - 8,
        day,
        hour,
        visits,
    }
}

function hideTooltip() {
    tooltip = null
}
</script>

<div class="heatmap-wrapper">
    {#if loading}
        <div class="loading-placeholder" style="height: {svgHeight}px; border-radius: var(--radius-sm);"></div>
    {:else}
        <svg
            width="100%"
            viewBox="0 0 {svgWidth} {svgHeight}"
            class="heatmap-svg"
            aria-label="Activity heatmap"
        >
            <!-- Day labels (top) -->
            {#each DAY_LABELS as day, i}
                <text
                    x={cellX(0) - 4}
                    y={cellY(i) + CELL_SIZE / 2 + 1}
                    class="axis-label"
                    text-anchor="end"
                    dominant-baseline="middle"
                >{day}</text>
            {/each}

            <!-- Hour labels (every 3 hours) -->
            {#each Array.from({ length: 8 }, (_, i) => i * 3) as hour}
                <text
                    x={cellX(hour) + CELL_SIZE / 2}
                    y={LABEL_H - 4}
                    class="axis-label"
                    text-anchor="middle"
                >{formatHour(hour)}</text>
            {/each}

            <!-- Cells -->
            {#each Array.from({ length: 7 }, (_, day) => day) as day}
                {#each Array.from({ length: 24 }, (_, hour) => hour) as hour}
                    {@const visits = cellValue(day, hour)}
                    <rect
                        x={cellX(hour)}
                        y={cellY(day)}
                        width={CELL_SIZE}
                        height={CELL_SIZE}
                        rx="2"
                        class="cell"
                        class:has-visits={visits > 0}
                        opacity={visits > 0 ? cellOpacity(visits) : 1}
                        onmouseenter={(e) => showTooltip(e, day, hour, visits)}
                        onmouseleave={hideTooltip}
                        tabindex="-1"
                        role="gridcell"
                        aria-label="{DAY_LABELS[day]} {formatHour(hour)}: {visits} visits"
                    />
                {/each}
            {/each}
        </svg>

        <!-- Tooltip (rendered outside SVG for proper positioning) -->
        {#if tooltip}
            <div
                class="tooltip"
                style="left: {tooltip.x}px; top: {tooltip.y}px;"
            >
                <strong>{DAY_LABELS[tooltip.day]} {formatHour(tooltip.hour)}</strong>
                <span>{tooltip.visits.toLocaleString()} visits</span>
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
        stroke: var(--text-muted);
        stroke-width: 0.5;
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