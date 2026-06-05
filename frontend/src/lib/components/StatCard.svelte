<script lang="ts">
import { formatNumber } from '$lib/utils'

interface Props {
    label: string
    value: number | string | null
    loading?: boolean
    subtitle?: string
}

let { label, value, loading = false, subtitle }: Props = $props()
</script>

<div class="stat-card">
    <span class="stat-label">{label}</span>
    {#if loading}
        <div class="stat-value loading-placeholder" style="width: 5rem; height: 1.5rem;"></div>
    {:else}
        <span class="stat-value">{value !== null ? (typeof value === 'number' ? formatNumber(value) : value) : '—'}</span>
    {/if}
    {#if subtitle}
        <span class="stat-subtitle">{subtitle}</span>
    {/if}
</div>

<style>
    .stat-card {
        display: flex;
        flex-direction: column;
        gap: 2px;
        padding: var(--space-3) var(--space-4);
        background: var(--surface);
        border: 1px solid var(--border);
        border-radius: var(--radius);
        min-width: 120px;
        flex: 1;
    }

    .stat-label {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
        font-weight: 500;
    }

    .stat-value {
        font-size: 1.5rem;
        font-weight: 600;
        color: var(--text);
        line-height: 1.2;
        font-variant-numeric: tabular-nums;
    }

    .stat-subtitle {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }
</style>