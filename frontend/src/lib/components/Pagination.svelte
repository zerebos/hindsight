<script lang="ts">
interface Props {
    page: number       // 0-indexed
    pageSize: number
    total: number
    onpage: (page: number) => void
}

let { page, pageSize, total, onpage }: Props = $props()

const totalPages = $derived(Math.ceil(total / pageSize))
const start      = $derived(page * pageSize + 1)
const end        = $derived(Math.min((page + 1) * pageSize, total))
</script>

{#if total > 0}
<div class="pagination">
    <span class="pagination-info">
        {start.toLocaleString()}–{end.toLocaleString()} of {total.toLocaleString()}
    </span>

    <div class="pagination-controls">
        <button
            class="page-btn"
            disabled={page === 0}
            onclick={() => onpage(0)}
            aria-label="First page"
        >«</button>

        <button
            class="page-btn"
            disabled={page === 0}
            onclick={() => onpage(page - 1)}
            aria-label="Previous page"
        >‹</button>

        <span class="page-indicator">
            {page + 1} / {totalPages}
        </span>

        <button
            class="page-btn"
            disabled={page >= totalPages - 1}
            onclick={() => onpage(page + 1)}
            aria-label="Next page"
        >›</button>

        <button
            class="page-btn"
            disabled={page >= totalPages - 1}
            onclick={() => onpage(totalPages - 1)}
            aria-label="Last page"
        >»</button>
    </div>
</div>
{/if}

<style>
    .pagination {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--space-2) var(--space-3);
        border-top: 1px solid var(--border);
        flex-shrink: 0;
    }

    .pagination-info {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
    }

    .pagination-controls {
        display: flex;
        align-items: center;
        gap: var(--space-1);
    }

    .page-btn {
        padding: 2px 6px;
        background: var(--surface-2);
        border: 1px solid var(--border);
        border-radius: 3px;
        color: var(--text-muted);
        font-size: var(--font-size-sm);
        cursor: pointer;
        transition: background 0.1s, color 0.1s;
        line-height: 1.4;
    }

    .page-btn:hover:not(:disabled) {
        background: var(--surface-hover);
        color: var(--text);
    }

    .page-btn:disabled {
        opacity: 0.3;
        cursor: default;
    }

    .page-indicator {
        font-size: var(--font-size-xs);
        color: var(--text-muted);
        padding: 0 var(--space-2);
        font-variant-numeric: tabular-nums;
    }
</style>