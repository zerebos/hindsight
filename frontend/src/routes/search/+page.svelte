<script lang="ts">
import { SearchVisits } from '$hindsight/searchservice'
import { searchState, clearSearch } from '$lib/stores/search.svelte'
import { appState } from '$lib/stores/app.svelte'
import { nullStr, formatDate, truncateUrl, formatNumber } from '$lib/utils'
import TimeRangePicker from '$lib/components/TimeRangePicker.svelte'
import Pagination from '$lib/components/Pagination.svelte'
import type { DashboardFilter } from '$hindsight/internal/app/models'

// Cancel in-flight request when new search starts
let currentRequest: { cancel: () => void } | null = null
let debounceTimer: ReturnType<typeof setTimeout>

async function runSearch() {
    if (currentRequest) {
        currentRequest.cancel()
        currentRequest = null
    }

    searchState.loading = true
    searchState.error = null

    try {
        const req = SearchVisits(searchState.params)
        currentRequest = req
        const results = await req
        // Only update if this is still the current request
        if (currentRequest === req) {
            searchState.results = results
            currentRequest = null
        }
    } catch (err: any) {
        // Ignore cancellation errors
        if (err?.name !== 'AbortError' && !String(err).includes('cancel')) {
            searchState.error = String(err)
        }
    } finally {
        searchState.loading = false
    }
}

function scheduleSearch() {
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(runSearch, 300)
}

function onTextInput(e: Event) {
    searchState.params.Text = (e.target as HTMLInputElement).value
    searchState.params.Page = 0
    scheduleSearch()
}

function onDomainInput(e: Event) {
    searchState.params.Domain = (e.target as HTMLInputElement).value
    searchState.params.Page = 0
    scheduleSearch()
}

function onSortChange(e: Event) {
    searchState.params.SortBy = (e.target as HTMLSelectElement).value
    searchState.params.Page = 0
    runSearch()
}

function onTimeRangeChange(filter: DashboardFilter) {
    searchState.params.StartTime = filter.StartTime
    searchState.params.EndTime   = filter.EndTime
    searchState.params.Page = 0
    runSearch()
}

function onPageChange(page: number) {
    searchState.params.Page = page
    runSearch()
}

function filterByDomain(domain: string) {
    searchState.params.Domain = domain
    searchState.params.Page   = 0
    runSearch()
}

function clearDomainFilter() {
    searchState.params.Domain = ''
    searchState.params.Page   = 0
    runSearch()
}

function onClear() {
    clearTimeout(debounceTimer)
    clearSearch()
}

// Source label lookup from registered sources
function sourceLabel(sourceId: number): string {
    const src = appState.sources.find(s => s.ID === sourceId)
    return src ? nullStr(src.Label, src.Browser) : String(sourceId)
}

// Derived time range filter for the picker
const timeFilter = $derived({
    StartTime: searchState.params.StartTime,
    EndTime:   searchState.params.EndTime,
})

const hasFilters = $derived(
    searchState.params.Text !== '' ||
    searchState.params.Domain !== '' ||
    searchState.params.StartTime !== 0
)
</script>

<div class="search-page">
    <!-- Filter bar -->
    <div class="filter-bar">
        <div class="filter-row">
            <div class="search-input-wrap">
                <span class="search-icon">⌕</span>
                <input
                    class="input search-input"
                    type="text"
                    placeholder="Search URLs and titles..."
                    value={searchState.params.Text}
                    oninput={onTextInput}
                    autocomplete="off"
                    spellcheck="false"
                />
                {#if searchState.params.Text}
                    <button class="clear-btn" onclick={() => {
                        searchState.params.Text = ''
                        searchState.params.Page = 0
                        runSearch()
                    }}>✕</button>
                {/if}
            </div>

            <div class="domain-input-wrap">
                <input
                    class="input domain-input"
                    type="text"
                    placeholder="Domain filter..."
                    value={searchState.params.Domain}
                    oninput={onDomainInput}
                    autocomplete="off"
                    spellcheck="false"
                />
                {#if searchState.params.Domain}
                    <button class="clear-btn" onclick={clearDomainFilter}>✕</button>
                {/if}
            </div>

            <select
                class="input sort-select"
                value={searchState.params.SortBy}
                onchange={onSortChange}
            >
                <option value="recent">Most recent</option>
                <option value="visits">Most visited</option>
            </select>

            <TimeRangePicker
                filter={timeFilter}
                onchange={onTimeRangeChange}
            />

            {#if hasFilters}
                <button class="btn clear-all-btn" onclick={onClear}>
                    Clear
                </button>
            {/if}
        </div>

        <!-- Active domain filter badge -->
        {#if searchState.params.Domain}
            <div class="active-filters">
                <span class="filter-badge">
                    domain: {searchState.params.Domain}
                    <button onclick={clearDomainFilter}>✕</button>
                </span>
            </div>
        {/if}
    </div>

    <!-- Results area -->
    <div class="results-area">
        {#if searchState.error}
            <div class="results-error">{searchState.error}</div>

        {:else if !searchState.results && !searchState.loading}
            <div class="results-empty">
                <p>Enter a search query or filter to explore your history</p>
                <p class="text-faint text-sm">
                    Searching across {formatNumber(appState.sources.reduce((n) => n, 0))} sources
                </p>
            </div>

        {:else if searchState.loading && !searchState.results}
            <div class="results-loading">
                {#each Array(12) as _}
                    <div class="result-skeleton">
                        <div class="loading-placeholder" style="height: 13px; width: 60%; border-radius: 3px;"></div>
                        <div class="loading-placeholder" style="height: 11px; width: 80%; margin-top: 4px; border-radius: 3px;"></div>
                    </div>
                {/each}
            </div>

        {:else if searchState.results}
            {#if searchState.results.Visits.length === 0}
                <div class="results-empty">
                    <p>No results found</p>
                    {#if hasFilters}
                        <button class="btn" onclick={onClear}>Clear filters</button>
                    {/if}
                </div>
            {:else}
                <div class="results-table-wrap" class:loading={searchState.loading}>
                    <table class="data-table results-table">
                        <thead>
                            <tr>
                                <th>Page</th>
                                <th>Domain</th>
                                <th>Visited</th>
                                <!-- <th style="text-align: right;">Visits</th> -->
                                <th>Source</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each searchState.results.Visits as visit}
                                {@const title = nullStr(visit.Title)}
                                <tr>
                                    <td class="page-cell" data-selectable>
                                        {#if title}
                                            <span class="visit-title">{title}</span>
                                        {/if}
                                        <span
                                            class="visit-url"
                                            class:no-title={!title}
                                            title={visit.Url}
                                        >
                                            {truncateUrl(visit.Url, 70)}
                                        </span>
                                    </td>
                                    <td class="domain-cell">
                                        <button
                                            class="domain-pill"
                                            onclick={() => filterByDomain(visit.Domain)}
                                            title="Filter by {visit.Domain}"
                                        >
                                            {visit.Domain}
                                        </button>
                                    </td>
                                    <td class="time-cell">
                                        {formatDate(visit.VisitedAt, 'datetime')}
                                    </td>
                                    <!-- <td class="count-cell">
                                        {visit.VisitCount > 1 ? visit.VisitCount : ''}
                                    </td> -->
                                    <td class="source-cell text-muted">
                                        {sourceLabel(visit.SourceID)}
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>

                <Pagination
                    page={searchState.results.Page}
                    pageSize={searchState.results.PageSize}
                    total={searchState.results.Total}
                    onpage={onPageChange}
                />
            {/if}
        {/if}
    </div>
</div>

<style>
    .search-page {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
    }

    /* Filter bar */
    .filter-bar {
        padding: var(--space-3) var(--space-4);
        border-bottom: 1px solid var(--border);
        background: var(--surface);
        flex-shrink: 0;
    }

    .filter-row {
        display: flex;
        align-items: center;
        gap: var(--space-2);
    }

    .search-input-wrap,
    .domain-input-wrap {
        position: relative;
        display: flex;
        align-items: center;
    }

    .search-input-wrap {
        flex: 1;
    }

    .search-icon {
        position: absolute;
        left: 8px;
        color: var(--text-faint);
        font-size: 1rem;
        pointer-events: none;
    }

    .search-input {
        padding-left: 28px;
    }

    .domain-input {
        width: 180px;
    }

    .clear-btn {
        position: absolute;
        right: 6px;
        background: none;
        border: none;
        color: var(--text-faint);
        font-size: 0.7rem;
        cursor: pointer;
        padding: 2px;
        line-height: 1;
    }

    .clear-btn:hover {
        color: var(--text-muted);
    }

    .sort-select {
        width: 130px;
        cursor: pointer;
    }

    .clear-all-btn {
        flex-shrink: 0;
        white-space: nowrap;
    }

    .active-filters {
        display: flex;
        gap: var(--space-2);
        margin-top: var(--space-2);
    }

    .filter-badge {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        padding: 2px 6px;
        background: var(--accent-subtle);
        border: 1px solid var(--accent);
        border-radius: 3px;
        font-size: var(--font-size-xs);
        color: var(--accent);
        font-family: var(--font-mono);
    }

    .filter-badge button {
        background: none;
        border: none;
        color: inherit;
        cursor: pointer;
        padding: 0;
        font-size: 0.65rem;
        line-height: 1;
    }

    /* Results */
    .results-area {
        flex: 1;
        overflow: hidden;
        display: flex;
        flex-direction: column;
    }

    .results-empty {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: var(--space-3);
        color: var(--text-muted);
        font-size: var(--font-size-sm);
        text-align: center;
    }

    .results-error {
        padding: var(--space-4);
        color: var(--error);
        font-size: var(--font-size-sm);
    }

    .results-loading {
        padding: var(--space-3) var(--space-4);
        display: flex;
        flex-direction: column;
        gap: var(--space-3);
    }

    .result-skeleton {
        padding: var(--space-2) 0;
        border-bottom: 1px solid var(--border-light);
    }

    .results-table-wrap {
        flex: 1;
        overflow-y: auto;
        transition: opacity 0.1s;
    }

    .results-table-wrap.loading {
        opacity: 0.5;
    }

    /* Table cells */
    .results-table :global(td) {
        vertical-align: top;
        padding-top: 7px;
        padding-bottom: 7px;
    }

    .page-cell {
        max-width: 0; /* forces truncation within table */
        width: 45%;
    }

    .visit-title {
        display: block;
        color: var(--text);
        font-size: var(--font-size-sm);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
    }

    .visit-url {
        display: block;
        font-family: var(--font-mono);
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 100%;
        margin-top: 1px;
    }

    .visit-url.no-title {
        color: var(--text);
        font-size: var(--font-size-sm);
    }

    .domain-cell {
        width: 15%;
        white-space: nowrap;
    }

    .domain-pill {
        background: none;
        border: none;
        color: var(--accent);
        font-family: var(--font-mono);
        font-size: var(--font-size-xs);
        cursor: pointer;
        padding: 1px 4px;
        border-radius: 3px;
        transition: background 0.1s;
        white-space: nowrap;
    }

    .domain-pill:hover {
        background: var(--accent-subtle);
    }

    .time-cell {
        width: 15%;
        white-space: nowrap;
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
    }

    .source-cell {
        width: 15%;
        font-size: var(--font-size-xs);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 120px;
    }
</style>