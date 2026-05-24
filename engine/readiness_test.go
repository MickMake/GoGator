package engine

import "testing"

func TestValidateEngineModeRejectsEmptyByDefault(t *testing.T) {
	pol := EngineModePolicy{RequireMinReadiness: true, MinReadiness: ShadowReadinessGoodMatch}
	sel, err := ValidateEngineMode(CandidateTripEvidence{}, ShadowSummary{Readiness: ShadowReadinessGoodMatch}, pol)
	if err != nil {
		t.Fatal(err)
	}
	if sel.Accepted {
		t.Fatalf("expected rejection")
	}
}

func TestValidateEngineModeInvalidMinReadiness(t *testing.T) {
	_, err := ValidateEngineMode(CandidateTripEvidence{}, ShadowSummary{}, EngineModePolicy{MinReadiness: ShadowReadiness("Bad")})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestValidateEngineModeIgnoresJitterOnlyUnmatchedLegacyForPercent(t *testing.T) {
	pol := EngineModePolicy{MinReadiness: ShadowReadinessNotEvaluated, MaxUnmatchedLegacyPercent: 0}
	summary := ShadowSummary{
		LegacyValidTripCount:          2,
		UnmatchedLegacyTripCount:      1,
		UnmatchedLegacyValidTripCount: 0,
	}
	sel, err := ValidateEngineMode(CandidateTripEvidence{Trips: []CandidateTrip{{}}}, summary, pol)
	if err != nil {
		t.Fatal(err)
	}
	if !sel.Accepted {
		t.Fatalf("expected accepted, got reasons=%v", sel.Reasons)
	}
}

func TestValidateEngineModeRejectsWhenValidLegacyUnmatchedPercentExceeded(t *testing.T) {
	pol := EngineModePolicy{MinReadiness: ShadowReadinessNotEvaluated, MaxUnmatchedLegacyPercent: 40}
	summary := ShadowSummary{
		LegacyValidTripCount:          2,
		UnmatchedLegacyTripCount:      2,
		UnmatchedLegacyValidTripCount: 1,
	}
	sel, err := ValidateEngineMode(CandidateTripEvidence{Trips: []CandidateTrip{{}}}, summary, pol)
	if err != nil {
		t.Fatal(err)
	}
	if sel.Accepted {
		t.Fatalf("expected rejection")
	}
}
