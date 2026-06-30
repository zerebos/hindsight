// ----------------------------------------------------------------
// Dashboard analytics
//
// Pure, dependency-free derivations computed on the frontend from the
// data the dashboard already fetches. No backend round-trips: every
// function here takes the existing query results and squeezes more
// insight out of them.
//
// Data shapes (from the Go bindings):
//   timeSeries: { Day: number (unix ms, UTC midnight), TotalVisits }[]
//   heatmap:    { Day: 0..6 (Sun..Sat), Hour: 0..23, TotalVisits }[]
//   topDomains: { Host, TotalVisits }[]
// ----------------------------------------------------------------

import { DAY_LABELS_FULL, formatHour } from './utils'

const DAY_MS = 86_400_000

export interface DayPoint {
    Day: number
    TotalVisits: number
}

export interface HeatCell {
    Day: number
    Hour: number
    TotalVisits: number
}

// ----------------------------------------------------------------
// Time series helpers
// ----------------------------------------------------------------

/**
 * Fill gaps in a daily series so every calendar day between the first
 * and last data point is present (missing days become 0). Produces the
 * continuous series an area/line chart needs to avoid misleading slopes
 * across missing days.
 */
export function fillDailyGaps(series: DayPoint[]): DayPoint[] {
    if (series.length === 0) return []
    const sorted = [...series].sort((a, b) => a.Day - b.Day)
    const byDay = new Map(sorted.map(d => [d.Day, d.TotalVisits]))
    const out: DayPoint[] = []
    const start = sorted[0].Day
    const end = sorted[sorted.length - 1].Day
    for (let day = start; day <= end; day += DAY_MS) {
        out.push({ Day: day, TotalVisits: byDay.get(day) ?? 0 })
    }
    return out
}

/** The single day with the most visits. */
export function busiestDay(series: DayPoint[]): DayPoint | null {
    if (series.length === 0) return null
    return series.reduce((best, d) => (d.TotalVisits > best.TotalVisits ? d : best))
}

/** Mean visits across active days (days that actually have visits). */
export function dailyAverage(series: DayPoint[]): number {
    const active = series.filter(d => d.TotalVisits > 0)
    if (active.length === 0) return 0
    const total = active.reduce((sum, d) => sum + d.TotalVisits, 0)
    return Math.round(total / active.length)
}

export interface Streaks {
    current: number
    longest: number
}

/**
 * Longest and current run of consecutive calendar days with activity.
 * "Current" is the run ending on the most recent active day in the data.
 */
export function streaks(series: DayPoint[]): Streaks {
    const days = series
        .filter(d => d.TotalVisits > 0)
        .map(d => d.Day)
        .sort((a, b) => a - b)
    if (days.length === 0) return { current: 0, longest: 0 }

    let longest = 1
    let run = 1
    for (let i = 1; i < days.length; i++) {
        if (days[i] - days[i - 1] === DAY_MS) {
            run++
        } else {
            run = 1
        }
        if (run > longest) longest = run
    }

    // Current streak: walk back from the most recent active day.
    let current = 1
    for (let i = days.length - 1; i > 0; i--) {
        if (days[i] - days[i - 1] === DAY_MS) current++
        else break
    }

    return { current, longest }
}

/**
 * Momentum: percentage change in visit volume between the first and
 * second half of the period. Positive means browsing is trending up.
 * Returns null when there isn't enough data to compare.
 */
export function momentum(series: DayPoint[]): number | null {
    if (series.length < 4) return null
    const mid = Math.floor(series.length / 2)
    const first = series.slice(0, mid).reduce((s, d) => s + d.TotalVisits, 0)
    const second = series.slice(mid).reduce((s, d) => s + d.TotalVisits, 0)
    if (first === 0) return second > 0 ? 100 : 0
    return Math.round(((second - first) / first) * 100)
}

// ----------------------------------------------------------------
// Heatmap aggregations
// ----------------------------------------------------------------

/** Total visits per hour-of-day (index 0..23). */
export function hourTotals(heatmap: HeatCell[]): number[] {
    const totals = new Array(24).fill(0)
    for (const c of heatmap) totals[c.Hour] += c.TotalVisits
    return totals
}

/** Total visits per day-of-week (index 0=Sun..6=Sat). */
export function weekdayTotals(heatmap: HeatCell[]): number[] {
    const totals = new Array(7).fill(0)
    for (const c of heatmap) totals[c.Day] += c.TotalVisits
    return totals
}

export interface Peak {
    index: number
    visits: number
    label: string
}

/** The hour of day with the most visits across the week. */
export function peakHour(heatmap: HeatCell[]): Peak | null {
    const totals = hourTotals(heatmap)
    return argMax(totals, formatHour)
}

/** The day of week with the most visits. */
export function peakWeekday(heatmap: HeatCell[]): Peak | null {
    const totals = weekdayTotals(heatmap)
    return argMax(totals, i => DAY_LABELS_FULL[i])
}

function argMax(totals: number[], label: (i: number) => string): Peak | null {
    let index = -1
    let visits = -1
    for (let i = 0; i < totals.length; i++) {
        if (totals[i] > visits) {
            visits = totals[i]
            index = i
        }
    }
    if (index < 0 || visits <= 0) return null
    return { index, visits, label: label(index) }
}

/** Share (0..1) of visits that land on Saturday or Sunday. */
export function weekendShare(heatmap: HeatCell[]): number {
    let weekend = 0
    let total = 0
    for (const c of heatmap) {
        total += c.TotalVisits
        if (c.Day === 0 || c.Day === 6) weekend += c.TotalVisits
    }
    return total > 0 ? weekend / total : 0
}

export interface TimeOfDay {
    label: string
    share: number
}

// Time-of-day blocks. Night wraps past midnight, so it's listed explicitly.
const TOD_BLOCKS: { label: string; hours: number[] }[] = [
    { label: 'Morning',   hours: [5, 6, 7, 8, 9, 10, 11] },
    { label: 'Afternoon', hours: [12, 13, 14, 15, 16] },
    { label: 'Evening',   hours: [17, 18, 19, 20, 21] },
    { label: 'Night',     hours: [22, 23, 0, 1, 2, 3, 4] },
]

/**
 * The part of the day a user browses most, with its share of visits.
 * A rough "chronotype" summary (Morning / Afternoon / Evening / Night).
 */
export function chronotype(heatmap: HeatCell[]): TimeOfDay | null {
    const hours = hourTotals(heatmap)
    const total = hours.reduce((s, n) => s + n, 0)
    if (total === 0) return null

    let best: TimeOfDay | null = null
    for (const block of TOD_BLOCKS) {
        const sum = block.hours.reduce((s, h) => s + hours[h], 0)
        const share = sum / total
        if (!best || share > best.share) best = { label: block.label, share }
    }
    return best
}

// ----------------------------------------------------------------
// Domain concentration
// ----------------------------------------------------------------

export interface DomainPoint {
    Host: string
    TotalVisits: number
}

export interface Concentration {
    topShare: number   // share of the single most-visited domain
    top3Share: number  // combined share of the top 3 domains
}

/**
 * How concentrated browsing is on a handful of domains, measured against
 * the period's total visit count.
 */
export function domainConcentration(domains: DomainPoint[], totalVisits: number): Concentration | null {
    if (totalVisits <= 0 || domains.length === 0) return null
    const top = domains[0]?.TotalVisits ?? 0
    const top3 = domains.slice(0, 3).reduce((s, d) => s + d.TotalVisits, 0)
    return {
        topShare: top / totalVisits,
        top3Share: top3 / totalVisits,
    }
}
