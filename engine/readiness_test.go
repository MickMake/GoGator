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

func TestValidateEngineModeRejectsWithoutSilentFallbackByDefault(t *testing.T) {
	pol := EngineModePolicy{
		MinReadiness:             ShadowReadinessNotEvaluated,
		AllowEmptyCandidates:     false,
		FallbackToLegacyOnReject: false,
	}
	sel, err := ValidateEngineMode(CandidateTripEvidence{}, ShadowSummary{Readiness: ShadowReadinessNotEvaluated}, pol)
	if err != nil {
		t.Fatal(err)
	}
	if sel.Accepted {
		t.Fatalf("expected rejection")
	}
	if sel.FallbackUsed {
		t.Fatalf("expected no fallback when fallback_to_legacy_on_reject=false")
	}
	if sel.SelectedTripSource != "engine" {
		t.Fatalf("expected selected_trip_source to remain engine on explicit reject, got %q", sel.SelectedTripSource)
	}
}

func TestValidateEngineModeUsesExplicitFallbackWhenEnabled(t *testing.T) {
	pol := EngineModePolicy{
		MinReadiness:             ShadowReadinessNotEvaluated,
		AllowEmptyCandidates:     false,
		FallbackToLegacyOnReject: true,
	}
	sel, err := ValidateEngineMode(CandidateTripEvidence{}, ShadowSummary{Readiness: ShadowReadinessNotEvaluated}, pol)
	if err != nil {
		t.Fatal(err)
	}
	if sel.Accepted {
		t.Fatalf("expected rejection")
	}
	if !sel.FallbackUsed {
		t.Fatalf("expected fallback when fallback_to_legacy_on_reject=true")
	}
	if sel.SelectedTripSource != "legacy" {
		t.Fatalf("expected selected_trip_source=legacy, got %q", sel.SelectedTripSource)
	}
}
