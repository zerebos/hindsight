<script lang="ts">
import { onMount } from 'svelte'
import { page } from '$app/state'
import { goto } from '$app/navigation'
import { Events } from '@wailsio/runtime'
import { GetSources } from '$hindsight/SourceService'
import { appState, markSyncStarted, markSyncComplete, markAllSyncsComplete, markSyncError } from '$lib/stores/app.svelte'
import { invalidateDashboard } from '$lib/stores/dashboard.svelte'
import type { SyncResult } from '$hindsight/internal/ingestion/models'
import '../app.css'
import { formatRelative } from '$lib/utils'

let { children } = $props()
let syncSettleTimer: ReturnType<typeof setTimeout>

// Current route for nav highlighting
// page from $app/state is already reactive in SvelteKit + Svelte 5

onMount(async () => {
    // Load registered sources on startup
    try {
        appState.sources = await GetSources()
    } catch (err) {
        appState.globalError = String(err)
    }

    // Wire up sync events from Go backend.
    // Wails wraps event data in { data: T } — access via .data
    Events.On('sync:started', () => {
        markSyncStarted()
    })

    // Track pending completions so we know when all sources are done
    let pendingResults: SyncResult[] = []

    Events.On('sync:complete', (event: { data: SyncResult }) => {
        pendingResults.push(event.data)
        markSyncComplete(event.data)
        invalidateDashboard()
        // Refresh sources list after sync in case labels changed
        GetSources().then(sources => { appState.sources = sources }).catch(() => {})
        // Mark all done — in practice SyncAll returns one result per source
        // and we don't know the total upfront, so we end syncing after
        // a short settling delay once results stop arriving
        clearTimeout(syncSettleTimer)
        syncSettleTimer = setTimeout(() => {
            markAllSyncsComplete()
            pendingResults = []
        }, 500)
    })

    Events.On('sync:error', (event: { data: string }) => {
        markSyncError(event.data)
    })

    appState.initialized = true
})

// Redirect to onboarding if no sources registered,
// or away from onboarding if sources already exist
$effect(() => {
    if (!appState.initialized) return
    const onOnboarding = page.url.pathname === '/onboarding'
    if (appState.sources.length === 0 && !onOnboarding) {
        goto('/onboarding')
    } else if (appState.sources.length > 0 && onOnboarding) {
        goto('/dashboard')
    }
})

const navItems = [
    { path: '/',         label: 'Dashboard', icon: '◈' },
    { path: '/search',   label: 'Search',    icon: '⌕' },
    { path: '/settings', label: 'Settings',  icon: '⚙' },
]
</script>

<div class="shell">
    {#if page.url.pathname !== '/onboarding'}
    <nav class="sidebar">
        <div class="sidebar-header">
            <span class="app-name">Hindsight</span>
        </div>

        <ul class="nav-list">
            {#each navItems as item}
                <li>
                    <a
                        href={item.path}
                        class="nav-item"
                        class:active={page.url.pathname === item.path}
                    >
                        <span class="nav-icon">{item.icon}</span>
                        <span class="nav-label">{item.label}</span>
                    </a>
                </li>
            {/each}
        </ul>

        <div class="sidebar-footer">
            {#if appState.syncing}
                <span class="sync-indicator syncing">Syncing...</span>
            {:else if appState.lastSyncTime}
                <span class="sync-indicator idle">Synced {formatRelative(appState.lastSyncTime)}</span>
            {:else if appState.initialized}
                <span class="sync-indicator idle">Ready</span>
            {/if}
        </div>
    </nav>
    {/if}

    <main class="content">
        {#if appState.globalError}
            <div class="global-error" role="alert">
                <span>{appState.globalError}</span>
                <button onclick={() => appState.globalError = null}>✕</button>
            </div>
        {/if}

        {@render children()}
    </main>
</div>

<style>
    .shell {
        display: flex;
        height: 100vh;
        overflow: hidden;
    }

    .sidebar {
        width: 200px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        border-right: 1px solid var(--border);
        background: var(--surface);
    }

    .sidebar-header {
        padding: 1.25rem 1rem;
        border-bottom: 1px solid var(--border);
    }

    .app-name {
        font-weight: 600;
        font-size: 1rem;
        letter-spacing: 0.02em;
    }

    .nav-list {
        list-style: none;
        margin: 0;
        padding: 0.5rem 0;
        flex: 1;
    }

    .nav-item {
        display: flex;
        align-items: center;
        gap: 0.625rem;
        padding: 0.5rem 1rem;
        text-decoration: none;
        color: var(--text-muted);
        font-size: 0.875rem;
        transition: background 0.1s, color 0.1s;
    }

    .nav-item:hover {
        background: var(--surface-hover);
        color: var(--text);
    }

    .nav-item.active {
        color: var(--accent);
        background: var(--accent-subtle);
        font-weight: 500;
    }

    .nav-icon {
        width: 1rem;
        text-align: center;
        font-size: 0.9rem;
    }

    .sidebar-footer {
        padding: 0.75rem 1rem;
        border-top: 1px solid var(--border);
        font-size: 0.75rem;
    }

    .sync-indicator.syncing { color: var(--accent); }
    .sync-indicator.idle    { color: var(--text-muted); }

    .content {
        flex: 1;
        overflow: auto;
        display: flex;
        flex-direction: column;
    }

    .global-error {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.5rem 1rem;
        background: var(--error-subtle);
        color: var(--error);
        font-size: 0.875rem;
        border-bottom: 1px solid var(--error);
    }

    .global-error button {
        background: none;
        border: none;
        cursor: pointer;
        color: inherit;
        padding: 0.25rem;
    }
</style>