<script lang="ts">
import type { GetVisitTimeSeriesRow, GetTopDomainsRow, GetDashboardStatsRow } from '$dbgen/models'
import type { HeatmapCell } from '$db/models'
import {
    dailyAverage,
    busiestDay,
    streaks,
    peakHour,
    peakWeekday,
    weekendShare,
    chronotype,
    momentum,
    domainConcentration,
} from '$lib/analytics'
import { formatNumber, formatPercent, formatSigned } from '$lib/utils'

interface Props {
    timeSeries: GetVisitTimeSeriesRow[]
    heatmap: HeatmapCell[]
    topDomains: GetTopDomainsRow[]
    stats: GetDashboardStatsRow | null
    loading?: boolean
}

let { timeSeries, heatmap, topDomains, stats, loading = false }: Props = $props()

interface Insight {
    label: string
    value: string
    detail: string
    tone?: 'up' | 'down'
}

const insights = $derived.by<Insight[]>(() => {
    const series = timeSeries as { Day: number; TotalVisits: number }[]
    const out: Insight[] = []

    const avg = dailyAverage(series)
    if (avg > 0) out.push({ label: 'Daily average', value: formatNumber(avg), detail: 'visits per active day' })

    const busy = busiestDay(series)
    if (busy && busy.TotalVisits > 0) {
        out.push({
            label: 'Busiest day',
            value: new Date(busy.Day).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
            detail: `${formatNumber(busy.TotalVisits)} visits`,
        })
    }

    const ph = peakHour(heatmap)
    if (ph) out.push({ label: 'Peak hour', value: ph.label, detail: 'most active time of day' })

    const pw = peakWeekday(heatmap)
    if (pw) out.push({ label: 'Top weekday', value: pw.label, detail: 'busiest day of the week' })

    const st = streaks(series)
    if (st.longest > 0) out.push({ label: 'Longest streak', value: `${st.longest} ${st.longest === 1 ? 'day' : 'days'}`, detail: 'consecutive active days' })

    const ct = chronotype(heatmap)
    if (ct) out.push({ label: 'Chronotype', value: ct.label, detail: `${formatPercent(ct.share)} of visits` })

    const ws = weekendShare(heatmap)
    if (heatmap.length > 0) out.push({ label: 'Weekend share', value: formatPercent(ws), detail: 'of visits on Sat/Sun' })

    const conc = domainConcentration(topDomains as { Host: string; TotalVisits: number }[], stats?.TotalVisits ?? 0)
    if (conc) out.push({ label: 'Top-3 focus', value: formatPercent(conc.top3Share), detail: 'of visits from 3 domains' })

    const mo = momentum(series)
    if (mo !== null) out.push({ label: 'Momentum', value: formatSigned(mo), detail: 'vs first half of period', tone: mo >= 0 ? 'up' : 'down' })

    return out
})
</script>

{#if loading}
    <div class="insight-grid">
        {#each Array(8) as _, i (i)}
            <div class="insight-card">
                <div class="loading-placeholder" style="height: 11px; width: 70%; margin-bottom: 8px;"></div>
                <div class="loading-placeholder" style="height: 20px; width: 50%;"></div>
            </div>
        {/each}
    </div>
{:else if insights.length > 0}
    <div class="insight-grid">
        {#each insights as ins (ins.label)}
            <div class="insight-card">
                <span class="insight-label">{ins.label}</span>
                <span class="insight-value" class:up={ins.tone === 'up'} class:down={ins.tone === 'down'}>{ins.value}</span>
                <span class="insight-detail">{ins.detail}</span>
            </div>
        {/each}
    </div>
{/if}

<style>
    .insight-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
        gap: var(--space-3);
    }

    .insight-card {
        display: flex;
        flex-direction: column;
        gap: 2px;
        padding: var(--space-3);
        background: var(--surface);
        border: 1px solid var(--border);
        border-radius: var(--radius);
    }

    .insight-label {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
        font-weight: 500;
    }

    .insight-value {
        font-size: 1.25rem;
        font-weight: 600;
        color: var(--text);
        line-height: 1.25;
        font-variant-numeric: tabular-nums;
    }

    .insight-value.up { color: var(--success); }
    .insight-value.down { color: var(--error); }

    .insight-detail {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }
</style>
