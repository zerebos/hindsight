package db

import (
	"sort"

	dbgen "github.com/zerebos/hindsight/internal/db/generated"
	"github.com/zerebos/hindsight/internal/ingestion"
)

// TrackingParamCount is a tracking parameter and the number of visits
// (weighted by visit_count) whose original URL carried it.
type TrackingParamCount struct {
	Name  string
	Count int64
}

// TrackedDomainCount is a domain and the number of its visits that carried
// at least one tracking parameter.
type TrackedDomainCount struct {
	Host  string
	Count int64
}

// TrackingStats summarizes tracking-parameter exposure across a set of visits.
type TrackingStats struct {
	TrackedVisits int64                // visits (visit_count weighted) carrying >= 1 known tracker
	TopParams     []TrackingParamCount // most common tracking parameters, descending
	TopDomains    []TrackedDomainCount // domains with the most tracked visits, descending
}

// BucketTracking tallies tracking-parameter exposure from raw_url rows.
//
// A visit counts as "tracked" only when its raw_url actually contains a known
// tracking parameter. This matters because raw_url is preserved on any
// normalization change (fragment stripping, re-encoding), not only tracker
// removal -- so a raw_url row is not by itself evidence of tracking.
//
// Counts are weighted by visit_count. topN bounds the returned param and
// domain lists (<= 0 returns all).
func BucketTracking(rows []dbgen.GetTrackedVisitsRow, topN int) TrackingStats {
	paramCounts := make(map[string]int64)
	domainCounts := make(map[string]int64)
	var tracked int64

	for _, r := range rows {
		if !r.RawUrl.Valid {
			continue
		}
		params := ingestion.TrackingParamsIn(r.RawUrl.String)
		if len(params) == 0 {
			continue
		}
		tracked += r.VisitCount
		domainCounts[r.Host] += r.VisitCount
		for _, p := range params {
			paramCounts[p] += r.VisitCount
		}
	}

	return TrackingStats{
		TrackedVisits: tracked,
		TopParams:     topParamCounts(paramCounts, topN),
		TopDomains:    topDomainCounts(domainCounts, topN),
	}
}

func topParamCounts(counts map[string]int64, topN int) []TrackingParamCount {
	out := make([]TrackingParamCount, 0, len(counts))
	for name, c := range counts {
		out = append(out, TrackingParamCount{Name: name, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name // stable, deterministic ordering
	})
	if topN > 0 && len(out) > topN {
		out = out[:topN]
	}
	return out
}

func topDomainCounts(counts map[string]int64, topN int) []TrackedDomainCount {
	out := make([]TrackedDomainCount, 0, len(counts))
	for host, c := range counts {
		out = append(out, TrackedDomainCount{Host: host, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Host < out[j].Host
	})
	if topN > 0 && len(out) > topN {
		out = out[:topN]
	}
	return out
}
