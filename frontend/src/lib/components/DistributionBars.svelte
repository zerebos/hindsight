<script lang="ts">
import { formatNumber } from '$lib/utils'

interface Props {
    values: number[]
    labels: string[]
    /** Optional axis tick labels keyed by index; falls back to `labels`. */
    tickEvery?: number
    loading?: boolean
    height?: number
    /** Tooltip caption for each bar, e.g. "visits". */
    unit?: string
}

let { values, labels, tickEvery = 1, loading = false, height = 120, unit = 'visits' }: Props = $props()

const BOTTOM = 18

let width = $state(320)

const max = $derived(Math.max(1, ...values))
const innerH = $derived(height - BOTTOM)
const n = $derived(values.length)
const slot = $derived(n > 0 ? width / n : width)
const barW = $derived(Math.max(2, slot * 0.64))
const peakIndex = $derived(values.indexOf(Math.max(...values)))

function barH(v: number): number {
    return Math.max(v > 0 ? 2 : 0, (v / max) * innerH)
}

let hovered = $state<number | null>(null)
</script>

<div class="dist" style="height: {height}px;" bind:clientWidth={width}>
    {#if loading}
        <div class="loading-placeholder" style="height: 100%; border-radius: var(--radius-sm);"></div>
    {:else}
        <svg width={width} height={height} role="img" aria-label="Distribution chart">
            {#each values as v, i (i)}
                {@const x = i * slot + (slot - barW) / 2}
                {@const h = barH(v)}
                <rect
                    x={x}
                    y={innerH - h}
                    width={barW}
                    height={h}
                    rx="2"
                    class="bar"
                    class:peak={i === peakIndex && v > 0}
                    opacity={hovered === null || hovered === i ? 1 : 0.5}
                    onmouseenter={() => (hovered = i)}
                    onmouseleave={() => (hovered = null)}
                    role="presentation"
                />
                {#if i % tickEvery === 0}
                    <text x={i * slot + slot / 2} y={height - 5} text-anchor="middle" class="tick">{labels[i]}</text>
                {/if}
            {/each}
        </svg>

        {#if hovered !== null}
            <div class="tooltip" style="left: {(hovered + 0.5) * slot}px;" class:flip={hovered > n * 0.6}>
                <strong>{labels[hovered]}</strong>
                <span>{formatNumber(values[hovered])} {unit}</span>
            </div>
        {/if}
    {/if}
</div>

<style>
    .dist {
        position: relative;
        width: 100%;
    }

    svg {
        display: block;
        overflow: visible;
    }

    .bar {
        fill: var(--accent);
        opacity: 0.55;
        transition: opacity 0.1s;
        cursor: default;
    }

    .bar.peak {
        opacity: 0.95;
    }

    .tick {
        fill: var(--text-faint);
        font-size: 9px;
        font-family: var(--font-sans);
    }

    .tooltip {
        position: absolute;
        top: 0;
        transform: translateX(-50%);
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: var(--radius-sm);
        padding: 3px 7px;
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
        transform: translateX(-50%);
    }
</style>
