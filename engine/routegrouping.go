package engine

import "sort"

type RouteSimilarity struct {
	SharedCells, TotalCells int
	Score                   float64
}
type RouteGroupMatch struct {
	CandidateTripID int
	Similarity      RouteSimilarity
}
type RouteGroup struct {
	RouteGroupID int
	Signature    string
	TripCount    int
	TripIDs      []int
	Matches      []RouteGroupMatch
}

func buildRouteGroups(signatures []RouteSignature, enabled bool) []RouteGroup {
	if !enabled || len(signatures) == 0 {
		return nil
	}
	bySig := map[string][]RouteSignature{}
	for _, s := range signatures {
		if s.Signature == "" {
			continue
		}
		bySig[s.Signature] = append(bySig[s.Signature], s)
	}
	keys := make([]string, 0, len(bySig))
	for k := range bySig {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]RouteGroup, 0, len(keys))
	for i, sig := range keys {
		entries := bySig[sig]
		sort.Slice(entries, func(a, b int) bool { return entries[a].CandidateTripID < entries[b].CandidateTripID })
		g := RouteGroup{RouteGroupID: i + 1, Signature: sig, TripCount: len(entries)}
		for _, e := range entries {
			g.TripIDs = append(g.TripIDs, e.CandidateTripID)
			g.Matches = append(g.Matches, RouteGroupMatch{CandidateTripID: e.CandidateTripID, Similarity: RouteSimilarity{SharedCells: e.CellCount, TotalCells: e.CellCount, Score: 1}})
		}
		out = append(out, g)
	}
	return out
}
