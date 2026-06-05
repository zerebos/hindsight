<script lang="ts">
    import {GetSources, SyncAll} from "$bindings/github.com/zerebos/hindsight/sourceservice";
    import {invalidateDashboard} from "$lib/stores";
import { appState, markAllSyncsComplete, markSyncError, markSyncStarted } from '$lib/stores/app.svelte'
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

const label = $derived.by(() => {
    if (appState.globalError) return 'Sync error'
    if (appState.syncing) return 'Syncing...'
    if (relative) return `Synced ${relative}`
    if (appState.initialized) return 'Ready'
    return 'Initializing...'
})

async function refreshSources() {
    try {
        appState.sources = await GetSources()
    } catch {
        // Ignore errors here since we'll show them in the UI if sources fail to load
    }
}

async function syncAll() {
    markSyncStarted()
    try {
        const results = await SyncAll()
        for (const r of results) {
            if (r.Error) markSyncError(r.Error)
        }
        markAllSyncsComplete()
        invalidateDashboard()
        await refreshSources()
    } catch (err) {
        markSyncError(String(err))
    }
}

// Tooltip state
let tooltip = $state<{ x: number; y: number;} | null>(null)

function showTooltip(event: MouseEvent) {
    tooltip = {
        x: (event.target as SVGElement).getBoundingClientRect().x + 8,
        y: (event.target as SVGElement).getBoundingClientRect().y - 8,
    }
}

function hideTooltip() {
    tooltip = null
}
</script>


<div class="sync-status" class:syncing={appState.syncing} class:error={!!appState.globalError} class:idle={!appState.syncing && appState.initialized}>
    <div class="sync-indicator"></div>
    <span class="sync-label">
        {label}
    </span>
    <button
        class="sync-icon"
        aria-label="Sync now"
        onclick={() => appState.syncing || appState.globalError ? null : syncAll()}
        disabled={appState.syncing || !!appState.globalError}
        onmouseenter={showTooltip}
        onmouseleave={hideTooltip}
    >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 256 256">
            <path d="M224,48V96a8,8,0,0,1-8,8H168a8,8,0,0,1,0-16h28.69L182.06,73.37a79.56,79.56,0,0,0-56.13-23.43h-.45A79.52,79.52,0,0,0,69.59,72.71,8,8,0,0,1,58.41,61.27a96,96,0,0,1,135,.79L208,76.69V48a8,8,0,0,1,16,0ZM186.41,183.29a80,80,0,0,1-112.47-.66L59.31,168H88a8,8,0,0,0,0-16H40a8,8,0,0,0-8,8v48a8,8,0,0,0,16,0V179.31l14.63,14.63A95.43,95.43,0,0,0,130,222.06h.53a95.36,95.36,0,0,0,67.07-27.33,8,8,0,0,0-11.18-11.44Z" />
        </svg>
    </button>
</div>

{#if tooltip}
    <div
        class="tooltip"
        style="left: {tooltip.x}px; top: {tooltip.y}px;"
    >
        <strong>Sync Now</strong>
    </div>
{/if}

<style>

.sync-status {
    display: flex;
    align-items: center;
    font-size: 0.875rem;
    color: var(--text-muted);
    gap: 10px;
}

.sync-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--text-muted);
    display: flex;
    justify-content: center;
    align-items: center;
}

.syncing .sync-indicator {
    background: var(--accent);
    animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
    0%, 100% {
        opacity: 1;
    }
    50% {
        opacity: 0.5;
    }
}

.idle .sync-indicator {
    background: var(--success);
}

.error .sync-indicator {
    background: var(--error);
}

.sync-label {
    flex: 1;
}
.syncing .sync-label {
    color: var(--accent);
}
.idle .sync-label {
    color: var(--text-muted);
}

.sync-icon {
    background: none;
    border: none;
    padding: 0;
    margin-left: 8px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
}

.sync-icon:disabled {
    cursor: not-allowed;
    opacity: 0.5;
}

.sync-icon svg {
    height: 16px;
    width: 16px;
    color: var(--text-muted);
}

.syncing .sync-icon svg {
    animation: spin 1s linear infinite;
    color: var(--accent);
}

@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}


.tooltip {
    position: fixed;
    transform: translate(-50%, -100%);
    background: var(--surface-2);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 4px 8px;
    font-size: var(--font-size-xs);
    color: var(--text);
    pointer-events: none;
    white-space: nowrap;
    z-index: 100;
    display: flex;
    flex-direction: column;
    gap: 1px;
}
</style>