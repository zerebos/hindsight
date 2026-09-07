import {SvelteSet, SvelteMap} from 'svelte/reactivity';
import type {Source} from '$hindsight/internal/db/generated/models';
import type {SyncResult} from '$hindsight/internal/ingestion/models';

// ResolvedTheme is what actually gets rendered as a `.shell` class. The stored
// preference is a plain string ('default', 'light', 'dark', 'amoled', or a
// legacy/unknown value); only these three have CSS variables defined.
export type ResolvedTheme = 'light' | 'dark' | 'amoled';

function prefersDark(): boolean {
    return typeof window !== 'undefined'
        && typeof window.matchMedia === 'function'
        && window.matchMedia('(prefers-color-scheme: dark)').matches;
}

// followsOS is true for any preference that isn't an explicit concrete theme —
// 'default', the legacy 'system', '', or anything unrecognized — all of which
// track the OS color scheme rather than a fixed appearance.
function followsOS(pref: string): boolean {
    return pref !== 'light' && pref !== 'dark' && pref !== 'amoled';
}

// resolveTheme maps a stored preference to a concrete rendered theme, so the UI
// never ends up with no theme class (and thus no CSS variables) applied.
function resolveTheme(pref: string): ResolvedTheme {
    if (!followsOS(pref)) return pref as ResolvedTheme;
    return prefersDark() ? 'dark' : 'light';
}

export const appState = $state({
    initialized: false,
    globalError: null as string | null,
    sources: [] as Source[],
    syncing: false,   // true while any sync is in progress
    lastSyncResults: new SvelteMap<number, SyncResult>(),
    lastSyncTime: null as number | null, // unix ms of last completed sync
    theme: resolveTheme('default') as ResolvedTheme,
});

// The raw stored preference, tracked so the OS listener below knows whether to
// re-resolve when the system color scheme changes.
let themePreference = 'default';

// applyTheme records the stored preference and updates the rendered theme,
// resolving OS-following preferences against the current OS color scheme.
export function applyTheme(pref: string) {
    themePreference = pref;
    appState.theme = resolveTheme(pref);
}

// While the preference follows the OS, track live color-scheme changes.
if (typeof window !== 'undefined' && window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (followsOS(themePreference)) {
            appState.theme = resolveTheme(themePreference);
        }
    });
}

export function markSyncStarted() {
    appState.syncing = true;
}

export function markSyncComplete(result: SyncResult) {
    const id = Number(result.SourceID);
    appState.lastSyncResults.set(id, result);
    // Syncing ends when the last result arrives — checked by caller
}

export function markAllSyncsComplete() {
    appState.syncing = false;
    appState.lastSyncTime = Date.now();
}

export function markSyncError(error: string) {
    appState.syncing = false;
    appState.globalError = error;
}