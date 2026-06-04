<script lang="ts">
import { onMount } from 'svelte'
import { goto } from '$app/navigation'
import { DiscoverSources, RegisterSource, SyncAll } from '$hindsight/sourceservice'
import { appState, markSyncStarted, markAllSyncsComplete } from '$lib/stores/app.svelte'
import type { DetectedSource } from '$hindsight/internal/browser/models'

// ----------------------------------------------------------------
// State
// ----------------------------------------------------------------

type DetectedWithSelection = DetectedSource & { selected: boolean }

let detecting   = $state(true)
let detected    = $state<DetectedWithSelection[]>([])
let detectError = $state<string | null>(null)

let registering  = $state(false)
let registerError = $state<string | null>(null)

const selected     = $derived(detected.filter(d => d.selected))
const noneSelected = $derived(selected.length === 0)
const noneFound    = $derived(!detecting && detected.length === 0 && !detectError)

onMount(async () => {
    try {
        const found = await DiscoverSources()
        detected = found.map(d => ({ ...d, selected: true }))
    } catch (err) {
        detectError = String(err)
    } finally {
        detecting = false
    }
})

function toggleAll(checked: boolean) {
    detected = detected.map(d => ({ ...d, selected: checked }))
}

async function getStarted() {
    if (noneSelected) return

    registering   = true
    registerError = null

    try {
        // Register each selected source individually
        const registered = []
        for (const source of selected) {
            const src = await RegisterSource(source)
            registered.push(src)
        }

        // Update global sources state
        appState.sources = registered

        // Kick off background sync — don't await, let it run while dashboard loads
        markSyncStarted()
        SyncAll().then(results => {
            markAllSyncsComplete()
            // SyncResult error handling is best-effort on onboarding
        }).catch(() => markAllSyncsComplete())

        goto('/dashboard')
    } catch (err) {
        registerError = String(err)
        registering   = false
    }
}

// Group detected sources by browser for cleaner display
const byBrowser = $derived.by(() => {
    const groups = new Map<string, DetectedWithSelection[]>()
    for (const d of detected) {
        const list = groups.get(d.Browser) ?? []
        list.push(d)
        groups.set(d.Browser, list)
    }
    return groups
})
</script>

<div class="onboarding">
    <div class="onboarding-card">
        <!-- Header -->
        <div class="onboarding-header">
            <span class="onboarding-logo">◈</span>
            <h1>Hindsight</h1>
            <p class="onboarding-tagline">Personal browser history intelligence</p>
        </div>

        <!-- Body -->
        <div class="onboarding-body">
            {#if detecting}
                <div class="detecting-state">
                    <div class="spinner"></div>
                    <span>Scanning for browsers…</span>
                </div>

            {:else if detectError}
                <div class="state-message error">
                    <p>Failed to detect browsers</p>
                    <p class="error-detail">{detectError}</p>
                </div>

            {:else if noneFound}
                <div class="state-message">
                    <p>No supported browsers found on this machine.</p>
                    <p class="text-faint text-sm">
                        Supported: Chrome, Edge, Brave, Vivaldi, Firefox, Zen, LibreWolf, Floorp, Safari
                    </p>
                </div>

            {:else}
                <div class="source-list-header">
                    <span class="source-count">
                        Found {detected.length} profile{detected.length !== 1 ? 's' : ''}
                    </span>
                    <div class="select-all-controls">
                        <button class="text-btn" onclick={() => toggleAll(true)}>All</button>
                        <span class="text-muted">·</span>
                        <button class="text-btn" onclick={() => toggleAll(false)}>None</button>
                    </div>
                </div>

                <div class="source-list">
                    {#each [...byBrowser.entries()] as [browser, sources]}
                        <div class="browser-group">
                            <span class="browser-group-name">{browser}</span>
                            {#each sources as source}
                                <label class="source-row">
                                    <input
                                        type="checkbox"
                                        bind:checked={source.selected}
                                    />
                                    <div class="source-info">
                                        <span class="source-label">{source.Label}</span>
                                        <span class="source-path text-faint">{source.Path}</span>
                                    </div>
                                </label>
                            {/each}
                        </div>
                    {/each}
                </div>
            {/if}

            {#if registerError}
                <div class="state-message error" style="margin-top: 1rem;">
                    {registerError}
                </div>
            {/if}
        </div>

        <!-- Footer -->
        <div class="onboarding-footer">
            {#if !detecting && !noneFound}
                <p class="footer-note">
                    {selected.length} of {detected.length} selected
                    · History is read locally, nothing leaves your machine
                </p>
                <button
                    class="btn btn-primary get-started-btn"
                    disabled={noneSelected || registering}
                    onclick={getStarted}
                >
                    {registering ? 'Setting up…' : 'Get Started'}
                </button>
            {/if}

            {#if noneFound}
                <button class="btn" onclick={() => goto('/dashboard')}>
                    Skip for now
                </button>
            {/if}
        </div>
    </div>
</div>

<style>
    .onboarding {
        height: 100vh;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg);
    }

    .onboarding-card {
        width: 480px;
        background: var(--surface);
        border: 1px solid var(--border);
        border-radius: var(--radius-lg);
        display: flex;
        flex-direction: column;
        max-height: 80vh;
        overflow: hidden;
    }

    /* Header */
    .onboarding-header {
        padding: var(--space-8) var(--space-6) var(--space-4);
        text-align: center;
        border-bottom: 1px solid var(--border);
        flex-shrink: 0;
    }

    .onboarding-logo {
        font-size: 2rem;
        color: var(--accent);
        display: block;
        margin-bottom: var(--space-2);
    }

    .onboarding-header h1 {
        font-size: 1.5rem;
        font-weight: 600;
        letter-spacing: -0.01em;
    }

    .onboarding-tagline {
        font-size: var(--font-size-sm);
        color: var(--text-muted);
        margin-top: var(--space-1);
    }

    /* Body */
    .onboarding-body {
        flex: 1;
        overflow-y: auto;
        padding: var(--space-4) var(--space-4) 0;
    }

    .detecting-state {
        display: flex;
        align-items: center;
        gap: var(--space-3);
        padding: var(--space-4) 0;
        color: var(--text-muted);
        font-size: var(--font-size-sm);
    }

    .spinner {
        width: 16px;
        height: 16px;
        border: 2px solid var(--border);
        border-top-color: var(--accent);
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
        flex-shrink: 0;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    .state-message {
        padding: var(--space-4) 0;
        font-size: var(--font-size-sm);
        color: var(--text-muted);
        text-align: center;
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
    }

    .state-message.error {
        color: var(--error);
    }

    .error-detail {
        font-size: var(--font-size-xs);
        font-family: var(--font-mono);
        color: var(--text-faint);
    }

    /* Source list */
    .source-list-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: var(--space-3);
    }

    .source-count {
        font-size: var(--font-size-sm);
        color: var(--text-muted);
    }

    .select-all-controls {
        display: flex;
        align-items: center;
        gap: var(--space-2);
        font-size: var(--font-size-xs);
        color: var(--text-faint);
    }

    .text-btn {
        background: none;
        border: none;
        color: var(--accent);
        font-size: var(--font-size-xs);
        cursor: pointer;
        padding: 0;
        font-family: inherit;
    }

    .text-btn:hover {
        text-decoration: underline;
    }

    .source-list {
        display: flex;
        flex-direction: column;
        gap: var(--space-3);
        padding-bottom: var(--space-4);
    }

    .browser-group {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .browser-group-name {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
        text-transform: uppercase;
        letter-spacing: 0.06em;
        font-weight: 500;
        padding: 0 0 var(--space-1) 0;
    }

    .source-row {
        display: flex;
        align-items: flex-start;
        gap: var(--space-3);
        padding: var(--space-2) var(--space-3);
        border-radius: var(--radius-sm);
        cursor: pointer;
        transition: background 0.1s;
        user-select: none;
    }

    .source-row:hover {
        background: var(--surface-hover);
    }

    .source-row input[type="checkbox"] {
        margin-top: 2px;
        flex-shrink: 0;
        accent-color: var(--accent);
        width: 14px;
        height: 14px;
        cursor: pointer;
    }

    .source-info {
        display: flex;
        flex-direction: column;
        gap: 1px;
        min-width: 0;
    }

    .source-label {
        font-size: var(--font-size-sm);
        color: var(--text);
    }

    .source-path {
        font-size: var(--font-size-xs);
        font-family: var(--font-mono);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    /* Footer */
    .onboarding-footer {
        padding: var(--space-4) var(--space-6) var(--space-6);
        border-top: 1px solid var(--border);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--space-3);
        flex-shrink: 0;
    }

    .footer-note {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
        text-align: center;
    }

    .get-started-btn {
        width: 100%;
        justify-content: center;
        padding: var(--space-3);
        font-size: var(--font-size-base);
    }
</style>