<script lang="ts">
import { appState } from '$lib/stores/app.svelte'
import { formatRelative } from '$lib/utils'
import {onMount} from "svelte";


function relativeSync(time: number | null): string | null {
    if (!time) return null;
    return formatRelative(time)
}

let relative = $derived(relativeSync(appState.lastSyncTime));

onMount(() => {
    // Update relative time every minute while synced
    const interval = setInterval(() => {
        relative = relativeSync(appState.lastSyncTime)
    }, 60 * 1000)

    return () => clearInterval(interval)
})
</script>


{#if appState.syncing}
    <span class="sync-indicator syncing">Syncing...</span>
{:else if relative}
    <span class="sync-indicator idle">Synced {relative}</span>
{:else if appState.initialized}
    <span class="sync-indicator idle">Ready</span>
{/if}


<style>
    .sync-indicator.syncing { color: var(--accent); }
    .sync-indicator.idle    { color: var(--text-muted); }
</style>