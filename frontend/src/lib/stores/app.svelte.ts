import {SvelteSet, SvelteMap} from 'svelte/reactivity';
import type {Source} from '$hindsight/internal/db/generated/models';
import type {SyncResult} from '$hindsight/internal/ingestion/models';

// ThemePreference is what the user selects; 'default' follows the OS color
// scheme. ResolvedTheme is what actually gets rendered as a `.shell` class —
// 'default' is never rendered directly (there are no CSS variables for it), so
// it is resolved to 'light' or 'dark' via prefers-color-scheme.
export type ThemePreference = 'light' | 'dark' | 'default' | 'amoled';
export type ResolvedTheme = 'light' | 'dark' | 'amoled';

function prefersDark(): boolean {
    return typeof window !== 'undefined'
        && window.matchMedia?.('(prefers-color-scheme: dark)').matches;
}

function resolveTheme(pref: ThemePreference): ResolvedTheme {
    return pref === 'default' ? (prefersDark() ? 'dark' : 'light') : pref;
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

// The raw user preference, tracked so the OS listener below knows whether to
// re-resolve when the system color scheme changes.
let themePreference: ThemePreference = 'default';

// applyTheme records the user's preference and updates the rendered theme,
// resolving 'default' against the current OS color scheme.
export function applyTheme(pref: ThemePreference) {
    themePreference = pref;
    appState.theme = resolveTheme(pref);
}

// While the preference is 'default', follow live OS color-scheme changes.
if (typeof window !== 'undefined' && window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (themePreference === 'default') {
            appState.theme = resolveTheme('default');
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