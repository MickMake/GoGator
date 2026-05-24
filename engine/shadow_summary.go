package engine

import (
	"fmt"
	"math"
	"sort"
	"time"

	"gogator/internal/gps"
)

type ShadowSeverity string

const (
	ShadowSeverityInfo    ShadowSeverity = "Info"
	ShadowSeverityWarning ShadowSeverity = "Warning"
	ShadowSeverityMajor   ShadowSeverity = "Major"
)

type ShadowReadiness string

const (
	ShadowReadinessNotEvaluated ShadowReadiness = "NotEvaluated"
	ShadowReadinessPoorMatch    ShadowReadiness = "PoorMatch"
	ShadowReadinessPartialMatch ShadowReadiness = "PartialMatch"
	ShadowReadinessGoodMatch    ShadowReadiness = "GoodMatch"
	ShadowReadinessExcellent    ShadowReadiness = "ExcellentMatch"
)

type ShadowMetric struct {
	Name     string
	Value    string
	Severity ShadowSeverity
	Notes    string
}

type ShadowMismatch struct {
	ID             string
	Type           string
	Severity       ShadowSeverity
	LegacyStart    time.Time
	LegacyEnd      time.Time
	CandidateStart time.Time
	CandidateEnd   time.Time
	DeltaMinutes   float64
	Notes          string
}

type ShadowSummary struct {
	LegacyValidTripCount           int
	LegacyJitterTripCount          int
	CandidateTripCount             int
	ApproxMatchedTripCount         int
	UnmatchedLegacyTripCount       int
	UnmatchedCandidateTripCount    int
	AverageStartDeltaMinutes       float64
	AverageEndDeltaMinutes         float64
	LargestBoundaryDeltaMinutes    float64
	OriginDestinationMismatchCount int
	LowConfidenceCandidateCount    int
	GapAffectedCandidateCount      int
	NoiseAffectedCandidateCount    int
	Readiness                      ShadowReadiness
	Metrics                        []ShadowMetric
	Mismatches                     []ShadowMismatch
}

type ShadowConfig struct {
	Enabled                        bool
	SummaryEnabled                 bool
	MatchToleranceMinutes          float64
	GoodMatchThresholdPercent      float64
	ExcellentMatchThresholdPercent float64
	WarnOnMajorMismatch            bool
}

type legacyComparisonTrip struct {
	Trip gps.Trip
}

func compareCandidateTrips(candidates CandidateTripEvidence, legacyValid, legacyJitter []gps.Trip, cfg ShadowConfig) CandidateTripComparison {
	summary := buildShadowSummary(candidates, legacyValid, legacyJitter, cfg)
	cmp := CandidateTripComparison{
		CandidateTripCount:           summary.CandidateTripCount,
		LegacyValidTripCount:         summary.LegacyValidTripCount,
		LegacyJitterTripCount:        summary.LegacyJitterTripCount,
		ApproxMatchedTrips:           summary.ApproxMatchedTripCount,
		MajorTimeBoundaryDifferences: 0,
		SiteDifferenceCount:          summary.OriginDestinationMismatchCount,
	}
	for _, mm := range summary.Mismatches {
		if mm.Type == "BoundaryDelta" && mm.Severity != ShadowSeverityInfo {
			cmp.MajorTimeBoundaryDifferences++
		}
	}
	cmp.UnmatchedCandidateTrips = make([]int, summary.UnmatchedCandidateTripCount)
	cmp.UnmatchedLegacyTrips = make([]int, summary.UnmatchedLegacyTripCount)
	cmp.ShadowSummary = summary
	return cmp
}

func buildShadowSummary(candidates CandidateTripEvidence, legacyValid, legacyJitter []gps.Trip, cfg ShadowConfig) ShadowSummary {
	if cfg.MatchToleranceMinutes <= 0 {
		cfg.MatchToleranceMinutes = 20
	}
	if cfg.GoodMatchThresholdPercent <= 0 {
		cfg.GoodMatchThresholdPercent = 70
	}
	if cfg.ExcellentMatchThresholdPercent <= 0 {
		cfg.ExcellentMatchThresholdPercent = 90
	}
	s := ShadowSummary{LegacyValidTripCount: len(legacyValid), LegacyJitterTripCount: len(legacyJitter), CandidateTripCount: len(candidates.Trips), Readiness: ShadowReadinessNotEvaluated}
	tol := cfg.MatchToleranceMinutes
	ci := sortedCandidates(candidates.Trips)
	legacyCombined := flattenLegacyTrips(legacyValid, legacyJitter)
	li := sortedLegacyComparison(legacyCombined)
	usedLegacy := map[int]bool{}
	var totalStart, totalEnd float64
	for _, cx := range ci {
		c := candidates.Trips[cx]
		match := -1
		for _, lx := range li {
			if usedLegacy[lx] {
				continue
			}
			l := legacyCombined[lx].Trip
			sd := math.Abs(c.StartTime.Sub(l.Start).Minutes())
			ed := math.Abs(c.EndTime.Sub(l.End).Minutes())
			if sd <= tol && ed <= tol {
				match = lx
				break
			}
		}
		if match == -1 {
			s.UnmatchedCandidateTripCount++
			s.Mismatches = append(s.Mismatches, ShadowMismatch{ID: "candidate-unmatched", Type: "UnmatchedCandidate", Severity: ShadowSeverityWarning, CandidateStart: c.StartTime, CandidateEnd: c.EndTime, Notes: "no legacy trip within tolerance"})
			continue
		}
		usedLegacy[match] = true
		s.ApproxMatchedTripCount++
		l := legacyCombined[match].Trip
		sd := math.Abs(c.StartTime.Sub(l.Start).Minutes())
		ed := math.Abs(c.EndTime.Sub(l.End).Minutes())
		totalStart += sd
		totalEnd += ed
		if sd > s.LargestBoundaryDeltaMinutes {
			s.LargestBoundaryDeltaMinutes = sd
		}
		if ed > s.LargestBoundaryDeltaMinutes {
			s.LargestBoundaryDeltaMinutes = ed
		}
		if (c.OriginLabel != "" && l.DepartureSite != "" && c.OriginLabel != l.DepartureSite) || (c.DestinationLabel != "" && l.DestinationSite != "" && c.DestinationLabel != l.DestinationSite) {
			s.OriginDestinationMismatchCount++
			s.Mismatches = append(s.Mismatches, ShadowMismatch{ID: "od-mismatch", Type: "OriginDestinationMismatch", Severity: ShadowSeverityWarning, LegacyStart: l.Start, LegacyEnd: l.End, CandidateStart: c.StartTime, CandidateEnd: c.EndTime, Notes: "site labels differ"})
		}
		boundarySeverity := ShadowSeverityInfo
		if sd > tol*0.8 || ed > tol*0.8 {
			boundarySeverity = ShadowSeverityWarning
		}
		if sd > tol || ed > tol {
			boundarySeverity = ShadowSeverityMajor
		}
		if boundarySeverity != ShadowSeverityInfo {
			s.Mismatches = append(s.Mismatches, ShadowMismatch{ID: "boundary-delta", Type: "BoundaryDelta", Severity: boundarySeverity, LegacyStart: l.Start, LegacyEnd: l.End, CandidateStart: c.StartTime, CandidateEnd: c.EndTime, DeltaMinutes: math.Max(sd, ed), Notes: "trip boundary drift"})
		}
	}
	for i := range legacyCombined {
		if !usedLegacy[i] {
			s.UnmatchedLegacyTripCount++
			l := legacyCombined[i].Trip
			s.Mismatches = append(s.Mismatches, ShadowMismatch{ID: "legacy-unmatched", Type: "UnmatchedLegacy", Severity: ShadowSeverityMajor, LegacyStart: l.Start, LegacyEnd: l.End, Notes: "no candidate trip matched"})
		}
	}
	for _, c := range candidates.Trips {
		if c.Confidence == CandidateConfidenceLow {
			s.LowConfidenceCandidateCount++
		}
		for _, w := range c.Warnings {
			if w == "gap_affected" {
				s.GapAffectedCandidateCount++
			}
			if w == "noise_affected" {
				s.NoiseAffectedCandidateCount++
			}
		}
	}
	if s.ApproxMatchedTripCount > 0 {
		s.AverageStartDeltaMinutes = totalStart / float64(s.ApproxMatchedTripCount)
		s.AverageEndDeltaMinutes = totalEnd / float64(s.ApproxMatchedTripCount)
	}
	matchPercent := 0.0
	legacyComparisonCount := len(legacyCombined)
	if legacyComparisonCount > 0 {
		matchPercent = (float64(s.ApproxMatchedTripCount) / float64(legacyComparisonCount)) * 100
	}
	if cfg.Enabled && cfg.SummaryEnabled {
		s.Readiness = ShadowReadinessPoorMatch
		if s.LegacyValidTripCount == 0 && s.CandidateTripCount == 0 {
			s.Readiness = ShadowReadinessNotEvaluated
		} else if matchPercent >= cfg.ExcellentMatchThresholdPercent && s.UnmatchedCandidateTripCount == 0 && s.OriginDestinationMismatchCount == 0 {
			s.Readiness = ShadowReadinessExcellent
		} else if matchPercent >= cfg.GoodMatchThresholdPercent {
			s.Readiness = ShadowReadinessGoodMatch
		} else if matchPercent >= 40 {
			s.Readiness = ShadowReadinessPartialMatch
		}
		s.Metrics = buildShadowMetrics(s, cfg)
	}
	return s
}

func buildShadowMetrics(s ShadowSummary, cfg ShadowConfig) []ShadowMetric {
	return []ShadowMetric{{"legacy_valid_trip_count", fmt.Sprintf("%d", s.LegacyValidTripCount), ShadowSeverityInfo, ""}, {"legacy_jitter_trip_count", fmt.Sprintf("%d", s.LegacyJitterTripCount), ShadowSeverityInfo, ""}, {"candidate_trip_count", fmt.Sprintf("%d", s.CandidateTripCount), ShadowSeverityInfo, ""}, {"approx_matched_trip_count", fmt.Sprintf("%d", s.ApproxMatchedTripCount), ShadowSeverityInfo, ""}, {"unmatched_legacy_trip_count", fmt.Sprintf("%d", s.UnmatchedLegacyTripCount), ShadowSeverityWarning, ""}, {"unmatched_candidate_trip_count", fmt.Sprintf("%d", s.UnmatchedCandidateTripCount), ShadowSeverityWarning, ""}, {"average_start_delta_minutes", fmt.Sprintf("%.2f", s.AverageStartDeltaMinutes), ShadowSeverityInfo, ""}, {"average_end_delta_minutes", fmt.Sprintf("%.2f", s.AverageEndDeltaMinutes), ShadowSeverityInfo, ""}, {"largest_boundary_delta_minutes", fmt.Sprintf("%.2f", s.LargestBoundaryDeltaMinutes), ShadowSeverityInfo, ""}, {"origin_destination_mismatch_count", fmt.Sprintf("%d", s.OriginDestinationMismatchCount), ShadowSeverityWarning, ""}, {"low_confidence_candidate_count", fmt.Sprintf("%d", s.LowConfidenceCandidateCount), ShadowSeverityWarning, ""}, {"gap_affected_candidate_count", fmt.Sprintf("%d", s.GapAffectedCandidateCount), ShadowSeverityWarning, ""}, {"noise_affected_candidate_count", fmt.Sprintf("%d", s.NoiseAffectedCandidateCount), ShadowSeverityWarning, ""}, {"readiness", string(s.Readiness), ShadowSeverityInfo, ""}}
}
func sortedCandidates(c []CandidateTrip) []int {
	idx := make([]int, len(c))
	for i := range c {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool { return c[idx[i]].StartTime.Before(c[idx[j]].StartTime) })
	return idx
}
func sortedLegacy(l []gps.Trip) []int {
	idx := make([]int, len(l))
	for i := range l {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool { return l[idx[i]].Start.Before(l[idx[j]].Start) })
	return idx
}

func flattenLegacyTrips(legacyValid, legacyJitter []gps.Trip) []legacyComparisonTrip {
	combined := make([]legacyComparisonTrip, 0, len(legacyValid)+len(legacyJitter))
	for _, trip := range legacyValid {
		combined = append(combined, legacyComparisonTrip{Trip: trip})
	}
	for _, trip := range legacyJitter {
		combined = append(combined, legacyComparisonTrip{Trip: trip})
	}
	return combined
}

func sortedLegacyComparison(l []legacyComparisonTrip) []int {
	idx := make([]int, len(l))
	for i := range l {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool { return l[idx[i]].Trip.Start.Before(l[idx[j]].Trip.Start) })
	return idx
}
