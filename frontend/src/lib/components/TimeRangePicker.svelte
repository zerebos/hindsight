<script lang="ts">
import type { DashboardFilter } from '$app/models'

interface Props {
    filter: DashboardFilter
    onchange: (filter: DashboardFilter) => void
}

let { filter, onchange }: Props = $props()

const presets = [
    { label: 'All',  days: 0 },
    { label: '7d',   days: 7 },
    { label: '30d',  days: 30 },
    { label: '90d',  days: 90 },
    { label: '1y',   days: 365 },
]

function select(days: number) {
    if (days === 0) {
        onchange({ StartTime: 0, EndTime: 0 })
    } else {
        onchange({
            StartTime: Date.now() - days * 86_400_000,
            EndTime: 0,
        })
    }
}

function isActive(days: number): boolean {
    if (days === 0) return filter.StartTime === 0
    const expected = Date.now() - days * 86_400_000
    // Allow 60s tolerance for clock drift
    return Math.abs(filter.StartTime - expected) < 60_000
}
</script>

<div class="range-picker" role="group" aria-label="Time range">
    {#each presets as preset}
        <button
            class="preset-btn"
            class:active={isActive(preset.days)}
            onclick={() => select(preset.days)}
        >
            {preset.label}
        </button>
    {/each}
</div>

<style>
    .range-picker {
        display: flex;
        gap: 2px;
    }

    .preset-btn {
        padding: 3px 8px;
        background: var(--surface-2);
        border: 1px solid var(--border);
        color: var(--text-muted);
        font-size: var(--font-size-xs);
        font-family: inherit;
        cursor: pointer;
        transition: background 0.1s, color 0.1s, border-color 0.1s;
        border-radius: 3px;
    }

    .preset-btn:hover {
        background: var(--surface-hover);
        color: var(--text);
    }

    .preset-btn.active {
        background: var(--accent-subtle);
        border-color: var(--accent);
        color: var(--accent);
    }
</style>