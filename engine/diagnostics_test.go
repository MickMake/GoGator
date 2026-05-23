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
