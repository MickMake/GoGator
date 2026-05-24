package engine

import (
	"fmt"
	"strings"
)

type EngineModePolicy struct {
	RequireMinReadiness       bool
	MinReadiness              ShadowReadiness
	AllowLowConfidence        bool
	AllowGapAffected          bool
	AllowEmptyCandidates      bool
	FallbackToLegacyOnReject  bool
	MaxUnmatchedLegacyPercent float64
	MaxBoundaryDeltaMinutes   float64
}

type EngineModeSelection struct {
	RequestedTripSource string
	SelectedTripSource  string
	Accepted            bool
	FallbackUsed        bool
	Readiness           ShadowReadiness
	Reasons             []string
	WarningCount        int
	MajorMismatchCount  int
}

func ValidateEngineMode(candidates CandidateTripEvidence, summary ShadowSummary, policy EngineModePolicy) (EngineModeSelection, error) {
	sel := EngineModeSelection{RequestedTripSource: "engine", SelectedTripSource: "engine", Accepted: true, Readiness: summary.Readiness}
	if err := validateReadinessValue(policy.MinReadiness); err != nil {
		return sel, err
	}
	if len(candidates.Trips) == 0 && !policy.AllowEmptyCandidates {
		sel.Reasons = append(sel.Reasons, "empty candidate trip output")
	}
	if policy.RequireMinReadiness && readinessRank(summary.Readiness) < readinessRank(policy.MinReadiness) {
		sel.Reasons = append(sel.Reasons, fmt.Sprintf("readiness %s below required %s", summary.Readiness, policy.MinReadiness))
	}
	if summary.LegacyValidTripCount > 0 && policy.MaxUnmatchedLegacyPercent >= 0 {
		pct := float64(summary.UnmatchedLegacyTripCount) * 100.0 / float64(summary.LegacyValidTripCount)
		if pct > policy.MaxUnmatchedLegacyPercent {
			sel.Reasons = append(sel.Reasons, fmt.Sprintf("unmatched legacy trips %.2f%% exceeds %.2f%%", pct, policy.MaxUnmatchedLegacyPercent))
		}
	}
	if policy.MaxBoundaryDeltaMinutes > 0 && summary.LargestBoundaryDeltaMinutes > policy.MaxBoundaryDeltaMinutes {
		sel.Reasons = append(sel.Reasons, fmt.Sprintf("largest boundary delta %.2f minutes exceeds %.2f", summary.LargestBoundaryDeltaMinutes, policy.MaxBoundaryDeltaMinutes))
	}
	if !policy.AllowLowConfidence && summary.LowConfidenceCandidateCount > 0 {
		sel.Reasons = append(sel.Reasons, fmt.Sprintf("low confidence candidates: %d", summary.LowConfidenceCandidateCount))
	}
	if !policy.AllowGapAffected && summary.GapAffectedCandidateCount > 0 {
		sel.Reasons = append(sel.Reasons, fmt.Sprintf("gap affected candidates: %d", summary.GapAffectedCandidateCount))
	}
	for _, mm := range summary.Mismatches {
		if mm.Severity == ShadowSeverityMajor {
			sel.MajorMismatchCount++
		}
		if mm.Severity == ShadowSeverityWarning || mm.Severity == ShadowSeverityMajor {
			sel.WarningCount++
		}
	}
	if len(sel.Reasons) > 0 {
		sel.Accepted = false
		if policy.FallbackToLegacyOnReject {
			sel.SelectedTripSource = "legacy"
			sel.FallbackUsed = true
		}
	}
	return sel, nil
}

func validateReadinessValue(r ShadowReadiness) error {
	if readinessRank(r) < 0 {
		return fmt.Errorf("invalid engine.engine_mode.min_readiness %q: valid values are NotEvaluated, PoorMatch, PartialMatch, GoodMatch, ExcellentMatch", r)
	}
	return nil
}

func readinessRank(r ShadowReadiness) int {
	switch strings.TrimSpace(string(r)) {
	case string(ShadowReadinessNotEvaluated):
		return 0
	case string(ShadowReadinessPoorMatch):
		return 1
	case string(ShadowReadinessPartialMatch):
		return 2
	case string(ShadowReadinessGoodMatch):
		return 3
	case string(ShadowReadinessExcellent):
		return 4
	default:
		return -1
	}
}
