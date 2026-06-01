<script lang="ts">
import { onMount } from 'svelte'
import { page } from '$app/state'
import { Events } from '@wailsio/runtime'
import { GetSources } from '$hindsight/sourceservice'
import { appState, markSyncComplete, markSyncError, invalidateDashboard } from '$lib/stores'
import type { SyncResult } from '$hindsight/internal/ingestion/models'

let { children } = $props()

// Current route for nav highlighting
const currentPath = $derived(page.url.pathname)

onMount(async () => {
    // Load registered sources on startup
    try {
        appState.sources = await GetSources()
    } catch (err) {
        appState.globalError = String(err)
    }

    // Wire up sync events from Go backend.
    // Wails wraps event data in { data: T } — access via .data
    Events.On('sync:complete', (event: { data: SyncResult }) => {
        markSyncComplete(event.data)
        invalidateDashboard()
    })

    Events.On('sync:error', (event: { data: string }) => {
        markSyncError(0, event.data)
    })

    appState.initialized = true
})

const navItems = [
    { path: '/',         label: 'Dashboard', icon: '◈' },
    { path: '/search',   label: 'Search',    icon: '⌕' },
    { path: '/settings', label: 'Settings',  icon: '⚙' },
]
</script>

<div class="shell">
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
                        class:active={currentPath === item.path}
                    >
                        <span class="nav-icon">{item.icon}</span>
                        <span class="nav-label">{item.label}</span>
                    </a>
                </li>
            {/each}
        </ul>

        <div class="sidebar-footer">
            {#if appState.syncingSourceIds.size > 0}
                <span class="sync-indicator syncing">Syncing...</span>
            {:else if appState.initialized}
                <span class="sync-indicator idle">Ready</span>
            {/if}
        </div>
    </nav>

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