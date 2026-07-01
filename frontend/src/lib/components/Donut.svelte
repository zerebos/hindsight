<script lang="ts">
import { pie as d3pie, arc as d3arc } from 'd3-shape'
import { formatCompact, formatNumber, formatPercent } from '$lib/utils'

export interface Segment {
    label: string
    value: number
    color: string
}

interface Props {
    segments: Segment[]
    /** Big label shown in the donut hole (e.g. total). */
    centerValue?: string
    centerLabel?: string
    loading?: boolean
    size?: number
}

let { segments, centerValue, centerLabel, loading = false, size = 160 }: Props = $props()

const total = $derived(segments.reduce((s, d) => s + d.value, 0))

const radius = $derived(size / 2)
const thickness = $derived(Math.max(14, size * 0.16))

const arcs = $derived.by(() => {
    const layout = d3pie<Segment>()
        .value(d => d.value)
        .sort(null)
        .padAngle(0.02)(segments.filter(s => s.value > 0))
    const gen = d3arc<(typeof layout)[number]>()
        .innerRadius(radius - thickness)
        .outerRadius(radius)
        .cornerRadius(2)
    return layout.map(a => ({ d: gen(a) ?? '', data: a.data }))
})

let hovered = $state<number | null>(null)
</script>

<div class="donut" class:loading>
    {#if loading}
        <div class="loading-placeholder" style="width: {size}px; height: {size}px; border-radius: 50%;"></div>
    {:else if total === 0}
        <div class="empty" style="height: {size}px;">No data</div>
    {:else}
        <svg width={size} height={size} viewBox="0 0 {size} {size}" role="img" aria-label="Breakdown by share">
            <g transform="translate({radius},{radius})">
                {#each arcs as a, i (a.data.label)}
                    <path
                        d={a.d}
                        fill={a.data.color}
                        class="seg"
                        opacity={hovered === null || hovered === i ? 1 : 0.35}
                        onmouseenter={() => (hovered = i)}
                        onmouseleave={() => (hovered = null)}
                        role="presentation"
                    />
                {/each}
            </g>
            <text x={radius} y={radius - 4} text-anchor="middle" class="center-value">
                {hovered !== null ? formatPercent(arcs[hovered].data.value / total) : (centerValue ?? formatCompact(total))}
            </text>
            <text x={radius} y={radius + 14} text-anchor="middle" class="center-label">
                {hovered !== null ? arcs[hovered].data.label : (centerLabel ?? 'total')}
            </text>
        </svg>

        <ul class="legend">
            {#each segments.filter(s => s.value > 0) as seg, i (seg.label)}
                <li
                    class="legend-item"
                    class:dim={hovered !== null && hovered !== i}
                    onmouseenter={() => (hovered = i)}
                    onmouseleave={() => (hovered = null)}
                    role="presentation"
                >
                    <span class="dot" style="background: {seg.color};"></span>
                    <span class="legend-label">{seg.label}</span>
                    <span class="legend-value">{formatNumber(seg.value)}</span>
                </li>
            {/each}
        </ul>
    {/if}
</div>

<style>
    .donut {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--space-4);
    }

    svg {
        flex-shrink: 0;
    }

    .seg {
        cursor: default;
        transition: opacity 0.12s;
    }

    .center-value {
        fill: var(--text);
        font-size: 1.1rem;
        font-weight: 600;
        font-variant-numeric: tabular-nums;
    }

    .center-label {
        fill: var(--text-faint);
        font-size: 0.625rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .legend {
        list-style: none;
        display: flex;
        flex-direction: column;
        gap: 6px;
        min-width: 0;
        width: 100%;
    }

    .legend-item {
        display: flex;
        align-items: center;
        gap: var(--space-2);
        font-size: var(--font-size-sm);
        transition: opacity 0.12s;
    }

    .legend-item.dim {
        opacity: 0.4;
    }

    .dot {
        width: 9px;
        height: 9px;
        border-radius: 2px;
        flex-shrink: 0;
    }

    .legend-label {
        color: var(--text);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        flex: 1;
    }

    .legend-value {
        color: var(--text-muted);
        font-size: var(--font-size-xs);
        font-variant-numeric: tabular-nums;
        flex-shrink: 0;
    }

    .empty {
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--text-faint);
        font-size: var(--font-size-sm);
        min-width: 120px;
    }
</style>
