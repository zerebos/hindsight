import {SvelteSet, SvelteMap} from 'svelte/reactivity';
import type {Source} from '$hindsight/internal/db/generated/models';
import type {SyncResult} from '$hindsight/internal/ingestion/models';

export const appState = $state({
    initialized: false,
    globalError: null as string | null,
    sources: [] as Source[],
    syncingSourceIds: new SvelteSet<number>(),
    lastSyncResults: new SvelteMap<number, SyncResult>(),
});

export const isSyncing = () => appState.syncingSourceIds.size > 0;

export function markSyncStarted(sourceId: number) {
    appState.syncingSourceIds.add(sourceId);
}

export function markSyncComplete(result: SyncResult) {
    const id = Number(result.SourceID);
    appState.syncingSourceIds.delete(id);
    appState.lastSyncResults.set(id, result);
}

export function markSyncError(sourceId: number, error: string) {
    appState.syncingSourceIds.delete(sourceId);
    appState.globalError = error;
}