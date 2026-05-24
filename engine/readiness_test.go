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
