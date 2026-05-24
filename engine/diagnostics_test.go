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
		Evidence:       EvidenceSet{Points: []PointEvidence{{Index: 1, Quality: QualityScore{Reasons: []QualityReason{ReasonGoodSignalMix}}}}},
		Stays:          StayEvidence{Stays: []Stay{{Reasons: []StayReason{StayReasonLowSpeed}, PointIndexes: []int{1, 2}}}},
		Visits:         VisitEvidence{Visits: []Visit{{Reasons: []VisitReason{VisitReasonFromStayType}, PointIndexes: []int{2, 3}}}},
		Excursions:     ExcursionEvidence{Excursions: []Excursion{{Reasons: []ExcursionReason{ExcursionReasonGapInfluence}, PointIndexes: []int{3, 4}}}},
		CandidateTrips: CandidateTripEvidence{Trips: []CandidateTrip{{Reasons: []CandidateTripReason{CandidateReasonGapInfluence}, Warnings: []string{"gap_affected"}}}},
		TripComparison: CandidateTripComparison{
			UnmatchedCandidateTrips: []int{1},
			UnmatchedLegacyTrips:    []int{2},
			ShadowSummary: ShadowSummary{
				Metrics:    []ShadowMetric{{Name: "x", Value: "1"}},
				Mismatches: []ShadowMismatch{{ID: "m1"}},
			},
		},
	}
	d := r.Diagnostics()
	r.Evidence.Points[0].Quality.Reasons[0] = ReasonInvalidCoordinates
	r.Stays.Stays[0].Reasons[0] = StayReasonPoorQuality
	r.Stays.Stays[0].PointIndexes[0] = 99
	r.Visits.Visits[0].Reasons[0] = VisitReasonGapInfluence
	r.Visits.Visits[0].PointIndexes[0] = 99
	r.Excursions.Excursions[0].Reasons[0] = ExcursionReasonLowVisitConf
	r.Excursions.Excursions[0].PointIndexes[0] = 99
	r.CandidateTrips.Trips[0].Reasons[0] = CandidateReasonBoundaryNoise
	r.CandidateTrips.Trips[0].Warnings[0] = "changed"
	r.TripComparison.UnmatchedCandidateTrips[0] = 99
	r.TripComparison.UnmatchedLegacyTrips[0] = 99
	r.TripComparison.ShadowSummary.Metrics[0].Value = "2"
	r.TripComparison.ShadowSummary.Mismatches[0].ID = "m2"
	if d.Points[0].Quality.Reasons[0] != ReasonGoodSignalMix || d.Stays[0].PointIndexes[0] != 1 || d.TripComparison.UnmatchedLegacyTrips[0] != 2 || d.ShadowSummary.Metrics[0].Value != "1" {
		t.Fatalf("nested slices were not deep-copied: %+v", d)
	}
}
