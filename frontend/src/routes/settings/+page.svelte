<script lang="ts">
import { onMount } from 'svelte'
import { GetSettings, UpdateSettings } from '$hindsight/settingsservice'
import {
    GetSources,
    RemoveSource,
    DiscoverAndRegister,
    SyncAll,
    SyncSource,
} from '$hindsight/sourceservice'
import { Settings } from '$hindsight/internal/config/models'
import { appState, markSyncStarted, markAllSyncsComplete, markSyncError } from '$lib/stores/app.svelte'
import { invalidateDashboard } from '$lib/stores/dashboard.svelte'
import { nullStr, formatDate, formatRelative } from '$lib/utils'

// ----------------------------------------------------------------
// Settings state
// ----------------------------------------------------------------

let settings = $state<Settings | null>(null)
let settingsLoading = $state(true)
let settingsError   = $state<string | null>(null)
let savedIndicator  = $state(false)
let savedTimer: ReturnType<typeof setTimeout>

onMount(async () => {
    try {
        settings = await GetSettings()
    } catch (err) {
        settingsError = String(err)
    } finally {
        settingsLoading = false
    }
})

async function save() {
    if (!settings) return
    try {
        await UpdateSettings(settings)
        appState.theme = settings.General.Theme as 'light' | 'dark' | 'default'
        clearTimeout(savedTimer)
        savedIndicator = true
        savedTimer = setTimeout(() => savedIndicator = false, 2000)
    } catch (err) {
        settingsError = String(err)
    }
}

// ----------------------------------------------------------------
// Source management
// ----------------------------------------------------------------

let sourcesWorking   = $state(false)
let sourcesError     = $state<string | null>(null)
// Row ID pending removal confirmation (null = none)
let confirmRemove    = $state<number | null>(null)

async function refreshSources() {
    try {
        appState.sources = await GetSources()
    } catch (err) {
        sourcesError = String(err)
    }
}

async function discoverNew() {
    sourcesWorking = true
    sourcesError   = null
    try {
        await DiscoverAndRegister()
        await refreshSources()
    } catch (err) {
        sourcesError = String(err)
    } finally {
        sourcesWorking = false
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

async function syncOne(sourceId: number) {
    markSyncStarted()
    try {
        const result = await SyncSource(sourceId)
        if (result.Error) markSyncError(result.Error)
        else markAllSyncsComplete()
        invalidateDashboard()
        await refreshSources()
    } catch (err) {
        markSyncError(String(err))
    }
}

async function removeSource(id: number) {
    sourcesWorking = true
    sourcesError   = null
    try {
        await RemoveSource(id)
        confirmRemove = null
        await refreshSources()
    } catch (err) {
        sourcesError = String(err)
    } finally {
        sourcesWorking = false
    }
}

const syncIntervalOptions = [
    { value: 0,   label: 'Manual only' },
    { value: 15,  label: 'Every 15 minutes' },
    { value: 30,  label: 'Every 30 minutes' },
    { value: 60,  label: 'Every hour' },
]

function getRelativeTimes(): Record<number, string> {
    const rels: Record<number, string> = {}
    for (const src of appState.sources) {
        if (!src.LastSyncedAt.Valid) rels[src.ID] = 'never';
        else rels[src.ID] = formatRelative(src.LastSyncedAt.Int64);
    }
    return rels
}

let relatives = $derived.by(getRelativeTimes);

onMount(() => {
    // Update relative times every minute
    const refreshTimer = setInterval(() => {
        relatives = getRelativeTimes();
    }, 60 * 1000)
    return () => clearInterval(refreshTimer)
});
</script>

<div class="settings-page">

    <!-- ----------------------------------------------------------------
         Sources
    ---------------------------------------------------------------- -->
    <section class="settings-section">
        <div class="section-header">
            <div>
                <h2 class="section-title">Sources</h2>
                <p class="section-desc">Registered browser profiles that Hindsight reads history from.</p>
            </div>
            <div class="section-actions">
                <button
                    class="btn"
                    disabled={sourcesWorking}
                    onclick={discoverNew}
                >
                    Discover New
                </button>
                <button
                    class="btn"
                    disabled={sourcesWorking || appState.syncing}
                    onclick={syncAll}
                >
                    {appState.syncing ? 'Syncing…' : 'Sync All'}
                </button>
            </div>
        </div>

        {#if sourcesError}
            <div class="inline-error">{sourcesError}</div>
        {/if}

        {#if appState.sources.length === 0}
            <div class="empty-sources">
                <p>No sources registered.</p>
                <button class="btn btn-primary" onclick={discoverNew}>Discover Browsers</button>
            </div>
        {:else}
            <table class="data-table sources-table">
                <thead>
                    <tr>
                        <th>Browser</th>
                        <th>Profile</th>
                        <th>Last Synced</th>
                        <th>Last Visit</th>
                        <th>Status</th>
                        <th></th>
                    </tr>
                </thead>
                <tbody>
                    {#each appState.sources as source}
                        <tr class:error-row={source.LastError.Valid}>
                            <td class="browser-cell">
                                <span class="browser-name">{source.Browser}</span>
                            </td>
                            <td class="profile-cell text-muted">
                                {nullStr(source.Label, source.Profile)}
                            </td>
                            <td class="time-cell">
                                {relatives[source.ID]}
                            </td>
                            <td class="time-cell">
                                {source.LastVisitSeen.Valid
                                    ? formatDate(source.LastVisitSeen.Int64, 'date')
                                    : '—'}
                            </td>
                            <td class="status-cell">
                                {#if source.LastError.Valid}
                                    <span
                                        class="status-badge error"
                                        title={source.LastError.String}
                                    >Error</span>
                                {:else if source.LastSyncedAt.Valid}
                                    <span class="status-badge ok">OK</span>
                                {:else}
                                    <span class="status-badge pending">New</span>
                                {/if}
                            </td>
                            <td class="actions-cell">
                                {#if confirmRemove === source.ID}
                                    <span class="confirm-text">Remove?</span>
                                    <button
                                        class="action-btn danger"
                                        onclick={() => removeSource(source.ID)}
                                    >Yes</button>
                                    <button
                                        class="action-btn"
                                        onclick={() => confirmRemove = null}
                                    >No</button>
                                {:else}
                                    <button
                                        class="action-btn"
                                        disabled={appState.syncing}
                                        onclick={() => syncOne(source.ID)}
                                        title="Sync this source"
                                    >Sync</button>
                                    <button
                                        class="action-btn danger"
                                        onclick={() => confirmRemove = source.ID}
                                        title="Remove this source"
                                    >Remove</button>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        {/if}
    </section>

    <!-- ----------------------------------------------------------------
         Sync settings
    ---------------------------------------------------------------- -->
    <section class="settings-section">
        <div class="section-header">
            <div>
                <h2 class="section-title">Sync</h2>
                <p class="section-desc">Control how and when Hindsight reads new history.</p>
            </div>
            {#if savedIndicator}
                <span class="saved-indicator">Saved</span>
            {/if}
        </div>

        {#if settingsLoading}
            <div class="settings-loading">
                {#each Array(3) as _}
                    <div class="loading-placeholder" style="height: 36px; border-radius: var(--radius-sm);"></div>
                {/each}
            </div>
        {:else if settings}
            <div class="settings-rows">
                <div class="settings-row">
                    <div class="settings-row-label">
                        <span>Sync interval</span>
                        <span class="settings-row-desc">How often to automatically check for new history</span>
                    </div>
                    <select
                        class="input settings-select"
                        value={settings.Sync.IntervalMinutes}
                        onchange={(e) => {
                            settings!.Sync.IntervalMinutes = Number((e.target as HTMLSelectElement).value)
                            save()
                        }}
                    >
                        {#each syncIntervalOptions as opt}
                            <option value={opt.value}>{opt.label}</option>
                        {/each}
                    </select>
                </div>

                <div class="settings-row">
                    <div class="settings-row-label">
                        <span>Sync on launch</span>
                        <span class="settings-row-desc">Automatically sync when the app starts</span>
                    </div>
                    <label class="toggle">
                        <input
                            type="checkbox"
                            checked={settings.Sync.SyncOnLaunch}
                            onchange={(e) => {
                                settings!.Sync.SyncOnLaunch = (e.target as HTMLInputElement).checked
                                save()
                            }}
                        />
                        <span class="toggle-track"></span>
                    </label>
                </div>

                <div class="settings-row">
                    <div class="settings-row-label">
                        <span>Sync on wake</span>
                        <span class="settings-row-desc">Sync when the computer wakes from sleep</span>
                    </div>
                    <label class="toggle">
                        <input
                            type="checkbox"
                            checked={settings.Sync.SyncOnWake}
                            onchange={(e) => {
                                settings!.Sync.SyncOnWake = (e.target as HTMLInputElement).checked
                                save()
                            }}
                        />
                        <span class="toggle-track"></span>
                    </label>
                </div>
            </div>
        {/if}
    </section>

    <!-- ----------------------------------------------------------------
         General settings
    ---------------------------------------------------------------- -->
    <section class="settings-section">
        <div class="section-header">
            <div>
                <h2 class="section-title">General</h2>
                <p class="section-desc">Appearance and startup behavior.</p>
            </div>
            {#if savedIndicator}
                <span class="saved-indicator">Saved</span>
            {/if}
        </div>

        {#if settings}
            <div class="settings-rows">
                <div class="settings-row">
                    <div class="settings-row-label">
                        <span>Theme</span>
                        <span class="settings-row-desc">Color scheme for the interface</span>
                    </div>
                    <select
                        class="input settings-select"
                        value={settings.General.Theme}
                        onchange={(e) => {
                            settings!.General.Theme = (e.target as HTMLSelectElement).value
                            save()
                        }}
                    >
                        <option value="default">Default</option>
                        <option value="dark">Dark</option>
                        <option value="light">Light</option>
                    </select>
                </div>

                <div class="settings-row">
                    <div class="settings-row-label">
                        <span>Launch at login</span>
                        <span class="settings-row-desc">Start Hindsight automatically when you log in</span>
                    </div>
                    <label class="toggle">
                        <input
                            type="checkbox"
                            checked={settings.General.LaunchAtLogin}
                            onchange={(e) => {
                                settings!.General.LaunchAtLogin = (e.target as HTMLInputElement).checked
                                save()
                            }}
                        />
                        <span class="toggle-track"></span>
                    </label>
                </div>

                <div class="settings-row">
                    <div class="settings-row-label">
                        <span>Minimize to tray</span>
                        <span class="settings-row-desc">Keep running in the system tray when the window is closed</span>
                    </div>
                    <label class="toggle">
                        <input
                            type="checkbox"
                            checked={settings.General.MinimizeToTray}
                            onchange={(e) => {
                                settings!.General.MinimizeToTray = (e.target as HTMLInputElement).checked
                                save()
                            }}
                        />
                        <span class="toggle-track"></span>
                    </label>
                </div>
            </div>
        {/if}
    </section>

</div>

<style>
    .settings-page {
        display: flex;
        flex-direction: column;
        gap: var(--space-4);
        padding: var(--space-4) var(--space-6);
        overflow-y: auto;
    }

    /* Section */
    .settings-section {
        background: var(--surface);
        border: 1px solid var(--border);
        border-radius: var(--radius);
    }

    .section-header {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        padding: var(--space-3) var(--space-4);
        border-bottom: 1px solid var(--border);
    }

    .section-title {
        font-size: var(--font-size-sm);
        font-weight: 600;
        color: var(--text-muted);
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .section-desc {
        font-size: var(--font-size-xs);
        color: var(--text-faint);
        margin-top: 2px;
    }

    .section-actions {
        display: flex;
        gap: var(--space-2);
        flex-shrink: 0;
    }

    /* Sources table */
    .sources-table {
        width: 100%;
    }

    .browser-cell .browser-name {
        font-weight: 500;
        text-transform: capitalize;
    }

    .profile-cell {
        font-size: var(--font-size-xs);
        max-width: 180px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .time-cell {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        white-space: nowrap;
        font-variant-numeric: tabular-nums;
    }

    .status-cell { white-space: nowrap; }

    .status-badge {
        display: inline-block;
        padding: 1px 6px;
        border-radius: 3px;
        font-size: var(--font-size-xs);
        font-weight: 500;
    }

    .status-badge.ok {
        background: var(--success-subtle);
        color: var(--success);
    }

    .status-badge.error {
        background: var(--error-subtle);
        color: var(--error);
        cursor: help;
    }

    .status-badge.pending {
        background: var(--accent-subtle);
        color: var(--accent);
    }

    .error-row td {
        background: var(--error-subtle);
    }

    .actions-cell {
        text-align: right;
        white-space: nowrap;
    }

    .action-btn {
        padding: 2px 8px;
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: 3px;
        color: var(--text-muted);
        font-size: var(--font-size-xs);
        font-family: inherit;
        cursor: pointer;
        transition: background 0.1s, color 0.1s;
        margin-left: 4px;
    }

    .action-btn:hover:not(:disabled) {
        background: var(--surface-hover);
        color: var(--text);
    }

    .action-btn:disabled {
        opacity: 0.4;
        cursor: default;
    }

    .action-btn.danger:hover:not(:disabled) {
        background: var(--error-subtle);
        color: var(--error);
        border-color: var(--error);
    }

    .confirm-text {
        font-size: var(--font-size-xs);
        color: var(--error);
        margin-right: 4px;
    }

    .empty-sources {
        padding: var(--space-6);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--space-3);
        color: var(--text-muted);
        font-size: var(--font-size-sm);
    }

    .inline-error {
        padding: var(--space-2) var(--space-4);
        color: var(--error);
        font-size: var(--font-size-xs);
        background: var(--error-subtle);
        border-bottom: 1px solid var(--border);
    }

    /* Settings rows */
    .settings-rows {
        display: flex;
        flex-direction: column;
    }

    .settings-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--space-3) var(--space-4);
        border-bottom: 1px solid var(--border-light);
    }

    .settings-row:last-child {
        border-bottom: none;
    }

    .settings-row-label {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .settings-row-label span:first-child {
        font-size: var(--font-size-sm);
        color: var(--text);
    }

    .settings-row-desc {
        font-size: var(--font-size-xs) !important;
        color: var(--text-faint) !important;
    }

    .settings-select {
        width: 160px;
        flex-shrink: 0;
    }

    .settings-loading {
        padding: var(--space-3) var(--space-4);
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
    }

    /* Toggle */
    .toggle {
        position: relative;
        display: inline-flex;
        align-items: center;
        cursor: pointer;
        flex-shrink: 0;
    }

    .toggle input {
        opacity: 0;
        width: 0;
        height: 0;
        position: absolute;
    }

    .toggle-track {
        width: 36px;
        height: 20px;
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: 10px;
        transition: background 0.2s, border-color 0.2s;
        position: relative;
    }

    .toggle-track::after {
        content: '';
        position: absolute;
        left: 2px;
        top: 2px;
        width: 14px;
        height: 14px;
        background: var(--text-faint);
        border-radius: 50%;
        transition: transform 0.2s, background 0.2s;
    }

    .toggle input:checked + .toggle-track {
        background: var(--accent-subtle);
        border-color: var(--accent);
    }

    .toggle input:checked + .toggle-track::after {
        transform: translateX(16px);
        background: var(--accent);
    }

    /* Saved indicator */
    .saved-indicator {
        font-size: var(--font-size-xs);
        color: var(--success);
        align-self: center;
        flex-shrink: 0;
    }
</style>