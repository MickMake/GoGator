package engine

import "gogator/engine/mapmatch"

type Diagnostics struct {
	Points          []PointEvidence
	Motion          []MotionSample
	Stays           []Stay
	Visits          []Visit
	Excursions      []Excursion
	CandidateTrips  []CandidateTrip
	MapMatch        []CandidateTripMapMatchDiagnostic
	RouteSignatures []RouteSignature
	RouteGroups     []RouteGroup
	TripComparison  CandidateTripComparison
	ShadowSummary   ShadowSummary
	EngineSelection EngineModeSelection
}

func (r Result) Diagnostics() Diagnostics {
	return Diagnostics{
		Points:          clonePointEvidence(r.Evidence.Points),
		Motion:          append([]MotionSample(nil), r.Motion.Samples...),
		Stays:           cloneStays(r.Stays.Stays),
		Visits:          cloneVisits(r.Visits.Visits),
		Excursions:      cloneExcursions(r.Excursions.Excursions),
		CandidateTrips:  cloneCandidateTrips(r.CandidateTrips.Trips),
		MapMatch:        cloneMapMatchDiagnostics(r.MapMatchDiagnostics),
		RouteSignatures: cloneRouteSignatures(r.RouteSignatures),
		RouteGroups:     cloneRouteGroups(r.RouteGroups),
		TripComparison:  cloneTripComparison(r.TripComparison),
		ShadowSummary:   cloneShadowSummary(r.TripComparison.ShadowSummary),
		EngineSelection: EngineModeSelection{},
	}
}

func cloneMapMatchDiagnostics(in []CandidateTripMapMatchDiagnostic) []CandidateTripMapMatchDiagnostic {
	out := append([]CandidateTripMapMatchDiagnostic(nil), in...)
	for i := range out {
		out[i].WarningCodes = append([]string(nil), out[i].WarningCodes...)
		out[i].MatchedShape = append([]mapmatch.MatchedShapePoint(nil), out[i].MatchedShape...)
	}
	return out
}

func cloneRouteSignatures(in []RouteSignature) []RouteSignature {
	out := append([]RouteSignature(nil), in...)
	for i := range out {
		out[i].Cells = append([]RouteSignatureCell(nil), out[i].Cells...)
		out[i].Warnings = append([]RouteSignatureWarning(nil), out[i].Warnings...)
	}
	return out
}

func cloneRouteGroups(in []RouteGroup) []RouteGroup {
	out := append([]RouteGroup(nil), in...)
	for i := range out {
		out[i].TripIDs = append([]int(nil), out[i].TripIDs...)
		out[i].Matches = append([]RouteGroupMatch(nil), out[i].Matches...)
	}
	return out
}

func clonePointEvidence(in []PointEvidence) []PointEvidence {
	out := append([]PointEvidence(nil), in...)
	for i := range out {
		out[i].Quality.Reasons = append([]QualityReason(nil), out[i].Quality.Reasons...)
	}
	return out
}
func cloneStays(in []Stay) []Stay {
	out := append([]Stay(nil), in...)
	for i := range out {
		out[i].Reasons = append([]StayReason(nil), out[i].Reasons...)
		out[i].PointIndexes = append([]int(nil), out[i].PointIndexes...)
	}
	return out
}
func cloneVisits(in []Visit) []Visit {
	out := append([]Visit(nil), in...)
	for i := range out {
		out[i].Reasons = append([]VisitReason(nil), out[i].Reasons...)
		out[i].PointIndexes = append([]int(nil), out[i].PointIndexes...)
	}
	return out
}
func cloneExcursions(in []Excursion) []Excursion {
	out := append([]Excursion(nil), in...)
	for i := range out {
		out[i].Reasons = append([]ExcursionReason(nil), out[i].Reasons...)
		out[i].PointIndexes = append([]int(nil), out[i].PointIndexes...)
	}
	return out
}
func cloneCandidateTrips(in []CandidateTrip) []CandidateTrip {
	out := append([]CandidateTrip(nil), in...)
	for i := range out {
		out[i].Reasons = append([]CandidateTripReason(nil), out[i].Reasons...)
		out[i].Warnings = append([]string(nil), out[i].Warnings...)
	}
	return out
}
func cloneShadowSummary(in ShadowSummary) ShadowSummary {
	out := in
	out.Metrics = append([]ShadowMetric(nil), in.Metrics...)
	out.Mismatches = append([]ShadowMismatch(nil), in.Mismatches...)
	return out
}
func cloneTripComparison(in CandidateTripComparison) CandidateTripComparison {
	out := in
	out.UnmatchedCandidateTrips = append([]int(nil), in.UnmatchedCandidateTrips...)
	out.UnmatchedLegacyTrips = append([]int(nil), in.UnmatchedLegacyTrips...)
	out.ShadowSummary = cloneShadowSummary(in.ShadowSummary)
	return out
}
