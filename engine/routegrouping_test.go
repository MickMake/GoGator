package engine

import "testing"

func TestRouteGroupingIdenticalSignatures(t *testing.T) {
	sigs := []RouteSignature{{CandidateTripID: 1, Signature: "a", CellCount: 2}, {CandidateTripID: 2, Signature: "a", CellCount: 2}, {CandidateTripID: 3, Signature: "b", CellCount: 2}}
	g := buildRouteGroups(sigs, true)
	if len(g) != 2 || g[0].TripCount != 2 {
		t.Fatalf("expected grouped identical signatures")
	}
}
