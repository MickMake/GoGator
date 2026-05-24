package engine

type Diagnostics struct {
	Points         []PointEvidence
	Motion         []MotionSample
	Stays          []Stay
	Visits         []Visit
	Excursions     []Excursion
	CandidateTrips []CandidateTrip
	TripComparison CandidateTripComparison
	ShadowSummary  ShadowSummary
}

func (r Result) Diagnostics() Diagnostics {
	tripComparison := cloneCandidateTripComparison(r.TripComparison)
	shadowSummary := cloneShadowSummary(r.TripComparison.ShadowSummary)
	tripComparison.ShadowSummary = shadowSummary
	return Diagnostics{
		Points:         clonePointEvidence(r.Evidence.Points),
		Motion:         append([]MotionSample(nil), r.Motion.Samples...),
		Stays:          cloneStays(r.Stays.Stays),
		Visits:         cloneVisits(r.Visits.Visits),
		Excursions:     cloneExcursions(r.Excursions.Excursions),
		CandidateTrips: cloneCandidateTrips(r.CandidateTrips.Trips),
		TripComparison: tripComparison,
		ShadowSummary:  shadowSummary,
	}
}

func clonePointEvidence(in []PointEvidence) []PointEvidence {
	out := append([]PointEvidence(nil), in...)
	for i := range out {
		out[i].Quality.Reasons = append([]QualityReason(nil), in[i].Quality.Reasons...)
	}
	return out
}

func cloneStays(in []Stay) []Stay {
	out := append([]Stay(nil), in...)
	for i := range out {
		out[i].Reasons = append([]StayReason(nil), in[i].Reasons...)
		out[i].PointIndexes = append([]int(nil), in[i].PointIndexes...)
	}
	return out
}

func cloneVisits(in []Visit) []Visit {
	out := append([]Visit(nil), in...)
	for i := range out {
		out[i].Reasons = append([]VisitReason(nil), in[i].Reasons...)
		out[i].PointIndexes = append([]int(nil), in[i].PointIndexes...)
	}
	return out
}

func cloneExcursions(in []Excursion) []Excursion {
	out := append([]Excursion(nil), in...)
	for i := range out {
		out[i].Reasons = append([]ExcursionReason(nil), in[i].Reasons...)
		out[i].PointIndexes = append([]int(nil), in[i].PointIndexes...)
	}
	return out
}

func cloneCandidateTrips(in []CandidateTrip) []CandidateTrip {
	out := append([]CandidateTrip(nil), in...)
	for i := range out {
		out[i].Reasons = append([]CandidateTripReason(nil), in[i].Reasons...)
		out[i].Warnings = append([]string(nil), in[i].Warnings...)
	}
	return out
}

func cloneCandidateTripComparison(in CandidateTripComparison) CandidateTripComparison {
	out := in
	out.UnmatchedCandidateTrips = append([]int(nil), in.UnmatchedCandidateTrips...)
	out.UnmatchedLegacyTrips = append([]int(nil), in.UnmatchedLegacyTrips...)
	return out
}

func cloneShadowSummary(in ShadowSummary) ShadowSummary {
	out := in
	out.Metrics = append([]ShadowMetric(nil), in.Metrics...)
	out.Mismatches = append([]ShadowMismatch(nil), in.Mismatches...)
	return out
}
