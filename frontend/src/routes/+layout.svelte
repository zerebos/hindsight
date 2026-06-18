<script lang="ts">
import { onMount, type Snippet } from 'svelte'
import { page } from '$app/state'
import { goto } from '$app/navigation'
import { Events } from '@wailsio/runtime'
import { GetSources } from '$hindsight/sourceservice'
import { appState, markSyncStarted, markSyncComplete, markAllSyncsComplete, markSyncError } from '$lib/stores/app.svelte'
import { invalidateDashboard } from '$lib/stores/dashboard.svelte'
import type { SyncResult } from '$hindsight/internal/ingestion/models'
import '../app.css'
    import SyncStatus from "$lib/components/SyncStatus.svelte";

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
    { path: '/dashboard',         label: 'Dashboard', icon: DashIcon as Snippet },
    { path: '/search',   label: 'Search',    icon: SearchIcon as Snippet },
    { path: '/settings', label: 'Settings',  icon: SettingsIcon as Snippet },
]
</script>


{#snippet DashIcon()}
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 256 256">
    <path d="M216,40H136V24a8,8,0,0,0-16,0V40H40A16,16,0,0,0,24,56V176a16,16,0,0,0,16,16H79.36L57.75,219a8,8,0,0,0,12.5,10l29.59-37h56.32l29.59,37a8,8,0,1,0,12.5-10l-21.61-27H216a16,16,0,0,0,16-16V56A16,16,0,0,0,216,40Zm0,136H40V56H216V176ZM104,120v24a8,8,0,0,1-16,0V120a8,8,0,0,1,16,0Zm32-16v40a8,8,0,0,1-16,0V104a8,8,0,0,1,16,0Zm32-16v56a8,8,0,0,1-16,0V88a8,8,0,0,1,16,0Z" />
</svg>
{/snippet}

{#snippet SearchIcon()}
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 256 256">
        <path d="M32,64a8,8,0,0,1,8-8H216a8,8,0,0,1,0,16H40A8,8,0,0,1,32,64Zm8,72h72a8,8,0,0,0,0-16H40a8,8,0,0,0,0,16Zm88,48H40a8,8,0,0,0,0,16h88a8,8,0,0,0,0-16Zm109.66,13.66a8,8,0,0,1-11.32,0L206,177.36A40,40,0,1,1,217.36,166l20.3,20.3A8,8,0,0,1,237.66,197.66ZM184,168a24,24,0,1,0-24-24A24,24,0,0,0,184,168Z" />
    </svg>
{/snippet}

{#snippet SettingsIcon()}
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 256 256">
        <path d="M128,80a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Zm0,80a32,32,0,1,1,32-32A32,32,0,0,1,128,160Zm109.94-52.79a8,8,0,0,0-3.89-5.4l-29.83-17-.12-33.62a8,8,0,0,0-2.83-6.08,111.91,111.91,0,0,0-36.72-20.67,8,8,0,0,0-6.46.59L128,41.85,97.88,25a8,8,0,0,0-6.47-.6A112.1,112.1,0,0,0,54.73,45.15a8,8,0,0,0-2.83,6.07l-.15,33.65-29.83,17a8,8,0,0,0-3.89,5.4,106.47,106.47,0,0,0,0,41.56,8,8,0,0,0,3.89,5.4l29.83,17,.12,33.62a8,8,0,0,0,2.83,6.08,111.91,111.91,0,0,0,36.72,20.67,8,8,0,0,0,6.46-.59L128,214.15,158.12,231a7.91,7.91,0,0,0,3.9,1,8.09,8.09,0,0,0,2.57-.42,112.1,112.1,0,0,0,36.68-20.73,8,8,0,0,0,2.83-6.07l.15-33.65,29.83-17a8,8,0,0,0,3.89-5.4A106.47,106.47,0,0,0,237.94,107.21Zm-15,34.91-28.57,16.25a8,8,0,0,0-3,3c-.58,1-1.19,2.06-1.81,3.06a7.94,7.94,0,0,0-1.22,4.21l-.15,32.25a95.89,95.89,0,0,1-25.37,14.3L134,199.13a8,8,0,0,0-3.91-1h-.19c-1.21,0-2.43,0-3.64,0a8.08,8.08,0,0,0-4.1,1l-28.84,16.1A96,96,0,0,1,67.88,201l-.11-32.2a8,8,0,0,0-1.22-4.22c-.62-1-1.23-2-1.8-3.06a8.09,8.09,0,0,0-3-3.06l-28.6-16.29a90.49,90.49,0,0,1,0-28.26L61.67,97.63a8,8,0,0,0,3-3c.58-1,1.19-2.06,1.81-3.06a7.94,7.94,0,0,0,1.22-4.21l.15-32.25a95.89,95.89,0,0,1,25.37-14.3L122,56.87a8,8,0,0,0,4.1,1c1.21,0,2.43,0,3.64,0a8.08,8.08,0,0,0,4.1-1l28.84-16.1A96,96,0,0,1,188.12,55l.11,32.2a8,8,0,0,0,1.22,4.22c.62,1,1.23,2,1.8,3.06a8.09,8.09,0,0,0,3,3.06l28.6,16.29A90.49,90.49,0,0,1,222.9,142.12Z" />
    </svg>
{/snippet}

<div class="shell" class:theme-light={appState.theme === 'light'} class:theme-dark={appState.theme === 'dark'} class:theme-amoled={appState.theme === 'amoled'}>
    {#if page.url.pathname !== '/onboarding'}
    <nav class="sidebar">
        <div class="sidebar-header">
            <span class="app-icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256">
                    <path d="M128,24a8,8,0,0,0-8,8V88a8,8,0,0,0,8,8,32,32,0,1,1-27.72,16,8,8,0,0,0-2.93-10.93l-48.5-28A8,8,0,0,0,37.92,76,104,104,0,1,0,128,24ZM48.09,91.1,83,111.26A48.09,48.09,0,0,0,80,128c0,1.53.08,3,.22,4.52L41.28,143A88.16,88.16,0,0,1,48.09,91.1Zm-2.67,67.31,39-10.44A48.1,48.1,0,0,0,120,175.32v40.31A88.2,88.2,0,0,1,45.42,158.41ZM136,215.63V175.32a48,48,0,0,0,0-94.65V40.36a88,88,0,0,1,0,175.27Z" />
                </svg>
            </span>
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
                        <span class="nav-icon">{@render item.icon()}</span>
                        <span class="nav-label">{item.label}</span>
                    </a>
                </li>
            {/each}
        </ul>

        <div class="sidebar-footer">
            <SyncStatus />
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
    .shell.theme-light {
        /* Base palette */
        --bg:           #f8f9fa;
        --surface:      #ffffff;
        --surface-2:    #f9fafb;
        --surface-hover:#f3f4f6;
        --border:       #e5e7eb;
        --border-light: #d1d5db;

        /* Text */
        --text:         #111111;
        --text-muted:   #6b7280;
        --text-faint:   #4a5568;

        /* Accent — blue, consistent with BD Blue family */
        --accent:        #4a8fe8;
        --accent-hover:  #5a9ef8;
        --accent-subtle: rgba(74, 143, 232, 0.08);
    }

    .shell.theme-dark {
        /* Base palette */
        --bg:           #111117;
        --surface:      #1e1e28;
        --surface-2:    #2a2a35;
        --surface-hover:#33333f;
        --border:       #3f3f4a;
        --border-light: #5c5c68;

        /* Text */
        --text:         #e5e7eb;
        --text-muted:   #9ca3af;
        --text-faint:   #6b7280;

        /* Accent — blue, consistent with BD Blue family */
        --accent:        #4a8fe8;
        --accent-hover:  #5a9ef8;
        --accent-subtle: rgba(74, 143, 232, 0.15);
    }

    .shell.theme-amoled {
        /* Base palette */
        --bg:           #000000;
        --surface:      #1a1a1a;
        --surface-2:    #2a2a2a;
        --surface-hover:#333333;
        --border:       #3f3f3f;
        --border-light: #5c5c5c;

        /* Text */
        --text:         #e5e7eb;
        --text-muted:   #9ca3af;
        --text-faint:   #6b7280;

        /* Accent — blue, consistent with BD Blue family */
        --accent:        #4a8fe8;
        --accent-hover:  #5a9ef8;
        --accent-subtle: rgba(74, 143, 232, 0.25);
    }

    .shell {
        color: var(--text);
        background: var(--bg);
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
        height: 60px;
        padding: 0.75rem;
        border-bottom: 1px solid var(--border);
        display: flex;
        align-items: center;
        gap: 0.375rem;
    }

    .app-icon {
        width: 32px;
        height: 32px;
        display: flex;
        align-items: center;
        justify-content: center;
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
        width: 24px;
        height: 24px;
        text-align: center;
        font-size: 0.9rem;
    }

    .nav-label {
        flex: 1;
        line-height: 1.715
    }

    .sidebar-footer {
        padding: 0.75rem 1rem;
        border-top: 1px solid var(--border);
        font-size: 0.75rem;
        height: 45px;
    }

    /* .sync-indicator.syncing { color: var(--accent); }
    .sync-indicator.idle    { color: var(--text-muted); } */

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