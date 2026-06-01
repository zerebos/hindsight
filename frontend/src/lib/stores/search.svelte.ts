import type {SearchParams, SearchResults} from '$hindsight/internal/app/models';

export const searchState = $state({
    params: {
        Text: '',
        Domain: '',
        StartTime: 0,
        EndTime: 0,
        SortBy: 'recent',
        Page: 0,
        PageSize: 50,
    } as SearchParams,
    results: null as SearchResults | null,
    loading: false,
    error: null as string | null,
});

export function clearSearch() {
    searchState.params = {
        Text: '',
        Domain: '',
        StartTime: 0,
        EndTime: 0,
        SortBy: 'recent',
        Page: 0,
        PageSize: 50,
    };
    searchState.results = null;
    searchState.error = null;
}