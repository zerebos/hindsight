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
    return n.toString();
}

// ----------------------------------------------------------------
// Day/time labels (for heatmap)
// ----------------------------------------------------------------

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