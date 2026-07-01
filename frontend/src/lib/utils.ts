import type {NullString, NullInt64} from '$bindings/database/sql/models';

// ----------------------------------------------------------------
// Nullable field helpers
// ----------------------------------------------------------------

/**
 * Unwrap a sql.NullString — returns the string value if valid,
 * otherwise returns the fallback (defaults to empty string).
 */
export function nullStr(n: NullString | null | undefined, fallback = ''): string {
    if (!n) return fallback;
    return n.Valid ? n.String : fallback;
}

/**
 * Unwrap a sql.NullInt64 — returns the number value if valid,
 * otherwise returns null.
 */
export function nullInt(n: NullInt64 | null | undefined): number | null {
    if (!n) return null;
    return n.Valid ? n.Int64 : null;
}

// ----------------------------------------------------------------
// Date and time formatting
// ----------------------------------------------------------------

/**
 * Format a unix millisecond timestamp to a human-readable string.
 * Returns 'never' for falsy values (0, null, undefined).
 */
export function formatDate(ms: number | null | undefined, format: 'date' | 'datetime' | 'relative' = 'datetime'): string {
    if (!ms) return 'never';
    if (format === 'relative') return formatRelative(ms);
    if (format === 'date') return new Date(ms).toLocaleDateString();
    return new Date(ms).toLocaleString();
}

/**
 * Format a unix millisecond timestamp as a relative time string.
 * e.g. "just now", "5m ago", "3h ago", "2d ago"
 */
export function formatRelative(ms: number): string {
    const diff = Date.now() - ms;
    const mins = Math.floor(diff / 60_000);
    const hours = Math.floor(diff / 3_600_000);
    const days = Math.floor(diff / 86_400_000);

    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return new Date(ms).toLocaleDateString();
}

type FormatDurationOptions = {
    /**
     * Use short unit labels (e.g., "d" instead of "days")
     * @default false
     */
    short?: boolean;
    /**
     * Maximum number of time units to display (e.g., 3 => "1 year, 2 months, 3 days")
     * @default 3
     */
    maxUnits?: number;
};

/**
 * Converts a duration in milliseconds to a human-readable string.
 *
 * @param ms - Duration in milliseconds (negative values are supported)
 * @param options - Formatting options
 * @returns Formatted duration string
 *
 * @example
 * formatDuration(93600000) // "1 day, 2 hours"
 * formatDuration(93600000, { short: true }) // "1d 2h"
 * formatDuration(7776000000, { maxUnits: 2 }) // "3 months"
 * formatDuration(-5000) // "minus 5 seconds"
 * formatDuration(0) // "0 seconds"
 */
export function formatDuration(
    ms: number,
    options: FormatDurationOptions = {}
): string {
    const {short = false, maxUnits = 3} = options;

    if (ms === 0) {
        return short ? "0s" : "0 seconds";
    }

    // Define units from largest to smallest
    const units = [
        {long: "year", short: "y", ms: 365 * 24 * 60 * 60 * 1000},
        {long: "month", short: "mo", ms: 30 * 24 * 60 * 60 * 1000},
        // {long: "week", short: "w", ms: 7 * 24 * 60 * 60 * 1000},
        {long: "day", short: "d", ms: 24 * 60 * 60 * 1000},
        {long: "hour", short: "h", ms: 60 * 60 * 1000},
        {long: "minute", short: "m", ms: 60 * 1000},
        {long: "second", short: "s", ms: 1000},
        {long: "millisecond", short: "ms", ms: 1},
    ];

    let remainder = Math.abs(ms);
    const parts: string[] = [];
    let unitsUsed = 0;

    for (const unit of units) {
        if (remainder >= unit.ms) {
            const count = Math.floor(remainder / unit.ms);
            remainder -= count * unit.ms;

            if (short) {
                parts.push(`${count}${unit.short}`);
            } else {
                const unitName = unit.long + (count !== 1 ? "s" : "");
                parts.push(`${count} ${unitName}`);
            }

            unitsUsed++;
            if (unitsUsed >= maxUnits || remainder === 0) {
                break;
            }
        }
    }

    let result = short ? parts.join(" ") : parts.join(", ");
    if (ms < 0) {
        result = `minus ${result}`;
    }

    return result;
}

/**
 * Format a unix millisecond timestamp as a short date string.
 * e.g. "2026-05-29"
 */
export function formatShortDate(ms: number): string {
    return new Date(ms).toISOString().slice(0, 10);
}

// ----------------------------------------------------------------
// Number formatting
// ----------------------------------------------------------------

/**
 * Format a large number with locale-appropriate separators.
 * e.g. 153946 -> "153,946"
 */
export function formatNumber(n: number): string {
    return n.toLocaleString();
}

/**
 * Format a number as a compact string for tight spaces.
 * e.g. 153946 -> "154K", 2531 -> "2.5K"
 */
export function formatCompact(n: number): string {
    if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
    if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
    return Math.round(n).toString();
}

/**
 * Format a fraction (0..1) as a rounded percentage string.
 * e.g. 0.423 -> "42%"
 */
export function formatPercent(fraction: number, digits = 0): string {
    return `${(fraction * 100).toFixed(digits)}%`;
}

/**
 * Format a signed percentage for trend/momentum display.
 * e.g. 12 -> "+12%", -5 -> "-5%"
 */
export function formatSigned(n: number, suffix = '%'): string {
    const sign = n > 0 ? '+' : '';
    return `${sign}${n}${suffix}`;
}

// ----------------------------------------------------------------
// Day/time labels (for heatmap)
// ----------------------------------------------------------------

// Categorical palette for charts (donut segments etc). Tuned to read well
// on both the dark and light themes. Index past the end wraps around.
export const CHART_PALETTE = [
    '#4a8fe8', // accent blue
    '#4ade80', // green
    '#fbbf24', // amber
    '#f472b6', // pink
    '#a78bfa', // violet
    '#22d3ee', // cyan
    '#fb923c', // orange
    '#94a3b8', // slate
] as const;

export function paletteColor(i: number): string {
    return CHART_PALETTE[i % CHART_PALETTE.length];
}

export const DAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] as const;
export const DAY_LABELS_FULL = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'] as const;

export function formatHour(hour: number): string {
    if (hour === 0) return '12am';
    if (hour < 12) return `${hour}am`;
    if (hour === 12) return '12pm';
    return `${hour - 12}pm`;
}

// ----------------------------------------------------------------
// URL helpers
// ----------------------------------------------------------------

/**
 * Truncate a URL for display, keeping the domain and trimming the path.
 * e.g. "https://github.com/zerebos/hindsight/pull/123" -> "github.com/zerebos/hindsight/pull/..."
 */
export function truncateUrl(url: string, maxLength = 60): string {
    try {
        const u = new URL(url);
        const display = u.host + u.pathname;
        if (display.length <= maxLength) return display;
        return display.slice(0, maxLength - 3) + '...';
    } catch {
        return url.length > maxLength ? url.slice(0, maxLength - 3) + '...' : url;
    }
}

/**
 * Extract the domain from a URL string.
 * Returns the full URL if parsing fails.
 */
export function extractDomain(url: string): string {
    try {
        return new URL(url).hostname.replace(/^www\./, '');
    } catch {
        return url;
    }
}