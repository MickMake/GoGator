package engine

import "testing"

func TestResultDiagnosticsCopiesSlices(t *testing.T) {
	r := Result{Evidence: EvidenceSet{Points: []PointEvidence{{Index: 1}}}, Motion: MotionEvidence{Samples: []MotionSample{{Index: 1}}}}
	d := r.Diagnostics()
	if len(d.Points) != 1 || len(d.Motion) != 1 {
		t.Fatalf("unexpected diagnostics: %+v", d)
	}
	r.Evidence.Points[0].Index = 99
	if d.Points[0].Index != 1 {
		t.Fatalf("expected copied points slice")
	}
}

func TestResultDiagnosticsDeepCopiesNestedSlices(t *testing.T) {
	r := Result{
		Stays: StayEvidence{Stays: []Stay{{Reasons: []StayReason{StayReasonLowSpeed}, PointIndexes: []int{1}}}},
		Visits: VisitEvidence{Visits: []Visit{{Reasons: []VisitReason{VisitReasonFromStayType}, PointIndexes: []int{2}}}},
		Excursions: ExcursionEvidence{Excursions: []Excursion{{Reasons: []ExcursionReason{ExcursionReasonVisitTransition}, PointIndexes: []int{3}}}},
		CandidateTrips: CandidateTripEvidence{Trips: []CandidateTrip{{Reasons: []CandidateTripReason{CandidateReasonVisitTransition}, Warnings: []string{"gap_affected"}}}},
		TripComparison: CandidateTripComparison{
			UnmatchedCandidateTrips: []int{4},
			UnmatchedLegacyTrips:    []int{5},
			ShadowSummary: ShadowSummary{
				Metrics:    []ShadowMetric{{Name: "readiness", Value: string(ShadowReadinessGoodMatch)}},
				Mismatches: []ShadowMismatch{{ID: "m1", Type: "BoundaryDelta"}},
			},
		},
	}
	d := r.Diagnostics()
	r.Stays.Stays[0].Reasons[0] = StayReasonPoorQuality
	r.Stays.Stays[0].PointIndexes[0] = 99
	r.Visits.Visits[0].Reasons[0] = VisitReasonGapInfluence
	r.Visits.Visits[0].PointIndexes[0] = 99
	r.Excursions.Excursions[0].Reasons[0] = ExcursionReasonGapInfluence
	r.Excursions.Excursions[0].PointIndexes[0] = 99
	r.CandidateTrips.Trips[0].Reasons[0] = CandidateReasonBoundaryNoise
	r.CandidateTrips.Trips[0].Warnings[0] = "noise_affected"
	r.TripComparison.UnmatchedCandidateTrips[0] = 99
	r.TripComparison.UnmatchedLegacyTrips[0] = 99
	r.TripComparison.ShadowSummary.Metrics[0].Value = string(ShadowReadinessPoorMatch)
	r.TripComparison.ShadowSummary.Mismatches[0].ID = "changed"

	if d.Stays[0].Reasons[0] != StayReasonLowSpeed || d.Stays[0].PointIndexes[0] != 1 {
		t.Fatalf("expected copied stay nested slices: %+v", d.Stays[0])
	}
	if d.Visits[0].Reasons[0] != VisitReasonFromStayType || d.Visits[0].PointIndexes[0] != 2 {
		t.Fatalf("expected copied visit nested slices: %+v", d.Visits[0])
	}
	if d.Excursions[0].Reasons[0] != ExcursionReasonVisitTransition || d.Excursions[0].PointIndexes[0] != 3 {
		t.Fatalf("expected copied excursion nested slices: %+v", d.Excursions[0])
	}
	if d.CandidateTrips[0].Reasons[0] != CandidateReasonVisitTransition || d.CandidateTrips[0].Warnings[0] != "gap_affected" {
		t.Fatalf("expected copied candidate trip nested slices: %+v", d.CandidateTrips[0])
	}
	if d.TripComparison.UnmatchedCandidateTrips[0] != 4 || d.TripComparison.UnmatchedLegacyTrips[0] != 5 {
		t.Fatalf("expected copied comparison nested slices: %+v", d.TripComparison)
	}
	if d.ShadowSummary.Metrics[0].Value != string(ShadowReadinessGoodMatch) || d.ShadowSummary.Mismatches[0].ID != "m1" {
		t.Fatalf("expected copied shadow summary nested slices: %+v", d.ShadowSummary)
	}
}
