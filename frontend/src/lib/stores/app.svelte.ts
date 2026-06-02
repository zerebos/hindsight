import {SvelteSet, SvelteMap} from 'svelte/reactivity';
import type {Source} from '$hindsight/internal/db/generated/models';
import type {SyncResult} from '$hindsight/internal/ingestion/models';

export const appState = $state({
    initialized: false,
    globalError: null as string | null,
    sources: [] as Source[],
    syncing: false,   // true while any sync is in progress
    lastSyncResults: new SvelteMap<number, SyncResult>(),
    lastSyncTime: null as number | null, // unix ms of last completed sync
});

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