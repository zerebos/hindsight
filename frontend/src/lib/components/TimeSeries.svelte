<script lang="ts">
import type { GetVisitTimeSeriesRow } from '$dbgen/models'
import { scaleTime, scaleLinear } from 'd3-scale'
import { line as d3line, area as d3area, curveMonotoneX } from 'd3-shape'
import { fillDailyGaps } from '$lib/analytics'
import { formatCompact, formatNumber } from '$lib/utils'

interface Props {
    data: GetVisitTimeSeriesRow[]
    loading?: boolean
    height?: number
}

let { data, loading = false, height = 220 }: Props = $props()

const M = { top: 12, right: 16, bottom: 24, left: 40 }

// Measured container width drives a fully responsive chart (no scroll).
let width = $state(640)

// Fill day gaps so the area doesn't draw misleading straight lines across
// days that simply had no visits.
const series = $derived(fillDailyGaps(data as { Day: number; TotalVisits: number }[]))

const innerW = $derived(Math.max(10, width - M.left - M.right))
const innerH = $derived(Math.max(10, height - M.top - M.bottom))

const x = $derived(
    scaleTime()
        .domain(
            series.length > 0
                ? [series[0].Day, series[series.length - 1].Day]
                : [0, 1],
        )
        .range([0, innerW]),
)

const yMax = $derived(Math.max(1, ...series.map(d => d.TotalVisits)))
const y = $derived(scaleLinear().domain([0, yMax]).nice().range([innerH, 0]))

const linePath = $derived(
    d3line<{ Day: number; TotalVisits: number }>()
        .x(d => x(d.Day))
        .y(d => y(d.TotalVisits))
        .curve(curveMonotoneX)(series) ?? '',
)

const areaPath = $derived(
    d3area<{ Day: number; TotalVisits: number }>()
        .x(d => x(d.Day))
        .y0(innerH)
        .y1(d => y(d.TotalVisits))
        .curve(curveMonotoneX)(series) ?? '',
)

const yTicks = $derived(y.ticks(4))

const xTicks = $derived.by(() => {
    const count = Math.max(2, Math.min(7, Math.floor(innerW / 90)))
    return x.ticks(count)
})

function fmtTick(d: Date): string {
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

// ---- Hover / crosshair ----
let hover = $state<{ i: number; cx: number } | null>(null)

function onMove(event: MouseEvent) {
    if (series.length === 0) return
    const svg = event.currentTarget as SVGRectElement
    const rect = svg.getBoundingClientRect()
    const px = ((event.clientX - rect.left) / rect.width) * innerW
    const t = x.invert(px).getTime()
    // Nearest day index.
    let best = 0
    let bestDist = Infinity
    for (let i = 0; i < series.length; i++) {
        const dist = Math.abs(series[i].Day - t)
        if (dist < bestDist) {
            bestDist = dist
            best = i
        }
    }
    hover = { i: best, cx: x(series[best].Day) }
}

function onLeave() {
    hover = null
}

const hoverPoint = $derived(hover ? series[hover.i] : null)
</script>

<div class="ts-wrap" style="height: {height}px;" bind:clientWidth={width}>
    {#if loading}
        <div class="loading-placeholder" style="height: 100%; border-radius: var(--radius-sm);"></div>
    {:else if series.length === 0}
        <div class="empty">No activity in this period</div>
    {:else}
        <svg width={width} height={height} class="ts-svg" role="img" aria-label="Visit history over time">
            <defs>
                <linearGradient id="ts-fill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%"   stop-color="var(--accent)" stop-opacity="0.35" />
                    <stop offset="100%" stop-color="var(--accent)" stop-opacity="0" />
                </linearGradient>
            </defs>

            <g transform="translate({M.left},{M.top})">
                <!-- Horizontal gridlines + y labels -->
                {#each yTicks as t (t)}
                    <line class="grid" x1="0" x2={innerW} y1={y(t)} y2={y(t)} />
                    <text class="axis-label y" x="-8" y={y(t)} dominant-baseline="middle" text-anchor="end">
                        {formatCompact(t)}
                    </text>
                {/each}

                <!-- x labels -->
                {#each xTicks as t (t.getTime())}
                    <text class="axis-label x" x={x(t)} y={innerH + 16} text-anchor="middle">
                        {fmtTick(t)}
                    </text>
                {/each}

                <!-- Area + line -->
                <path d={areaPath} fill="url(#ts-fill)" />
                <path d={linePath} class="line" />

                <!-- Hover crosshair -->
                {#if hover && hoverPoint}
                    <line class="crosshair" x1={hover.cx} x2={hover.cx} y1="0" y2={innerH} />
                    <circle class="dot" cx={hover.cx} cy={y(hoverPoint.TotalVisits)} r="3.5" />
                {/if}

                <!-- Capture layer -->
                <rect
                    x="0" y="0" width={innerW} height={innerH}
                    fill="transparent"
                    onmousemove={onMove}
                    onmouseleave={onLeave}
                    role="presentation"
                />
            </g>
        </svg>

        {#if hover && hoverPoint}
            <div
                class="tooltip"
                class:flip={hover.cx > innerW * 0.6}
                style="left: {M.left + hover.cx}px; top: {M.top}px;"
            >
                <strong>{new Date(hoverPoint.Day).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })}</strong>
                <span>{formatNumber(hoverPoint.TotalVisits)} visits</span>
            </div>
        {/if}
    {/if}
</div>

<style>
    .ts-wrap {
        position: relative;
        width: 100%;
    }

    .ts-svg {
        display: block;
        overflow: visible;
    }

    .line {
        fill: none;
        stroke: var(--accent);
        stroke-width: 2;
        stroke-linejoin: round;
        stroke-linecap: round;
    }

    .grid {
        stroke: var(--border-light);
        stroke-width: 1;
        shape-rendering: crispEdges;
        opacity: 0.6;
    }

    .axis-label {
        fill: var(--text-faint);
        font-size: 10px;
        font-family: var(--font-sans);
    }

    .crosshair {
        stroke: var(--text-faint);
        stroke-width: 1;
        stroke-dasharray: 3 3;
    }

    .dot {
        fill: var(--accent);
        stroke: var(--surface);
        stroke-width: 2;
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
        position: absolute;
        transform: translateX(8px);
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: var(--radius-sm);
        padding: 4px 8px;
        font-size: var(--font-size-xs);
        color: var(--text);
        pointer-events: none;
        white-space: nowrap;
        z-index: 10;
        display: flex;
        flex-direction: column;
        gap: 1px;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
    }

    .tooltip.flip {
        transform: translateX(-100%) translateX(-8px);
    }
</style>
