package engine

import (
	"time"

	"gogator/internal/gps"
)

type ExcursionType string
type ExcursionReason string
type ExcursionConfidence string

const (
	ExcursionBetweenKnownSites ExcursionType = "BetweenKnownSites"
	ExcursionKnownToUnknown    ExcursionType = "KnownSiteToUnknown"
	ExcursionUnknownToKnown    ExcursionType = "UnknownToKnownSite"
	ExcursionSupplierPickup    ExcursionType = "SupplierPickupCandidate"
	ExcursionShortOutAndBack   ExcursionType = "ShortOutAndBack"
	ExcursionLoop              ExcursionType = "LoopCandidate"
	ExcursionGapAffected       ExcursionType = "GapAffected"
	ExcursionLowConfidence     ExcursionType = "LowConfidence"
)
const (
	ExcursionReasonVisitTransition ExcursionReason = "visit_transition"
	ExcursionReasonShortReturn     ExcursionReason = "short_return"
	ExcursionReasonGapInfluence    ExcursionReason = "gap_influence"
	ExcursionReasonLowVisitConf    ExcursionReason = "low_visit_confidence"
)
const (
	ExcursionConfidenceLow    ExcursionConfidence = "Low"
	ExcursionConfidenceMedium ExcursionConfidence = "Medium"
	ExcursionConfidenceHigh   ExcursionConfidence = "High"
)

type Excursion struct {
	FromVisitIndex int
	ToVisitIndex   int
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	DistanceM      float64
	PointIndexes   []int
	Type           ExcursionType
	Reasons        []ExcursionReason
	Confidence     ExcursionConfidence
}
type ExcursionEvidence struct{ Excursions []Excursion }

type ExcursionConfig struct {
	Enabled                    bool
	ShortOutAndBackMaxMinutes  float64
	ShortOutAndBackMaxDistance float64
}

func detectExcursions(visits VisitEvidence, cfg ExcursionConfig) ExcursionEvidence {
	out := ExcursionEvidence{}
	if !cfg.Enabled || len(visits.Visits) < 2 {
		return out
	}
	if cfg.ShortOutAndBackMaxMinutes <= 0 {
		cfg.ShortOutAndBackMaxMinutes = 20
	}
	if cfg.ShortOutAndBackMaxDistance <= 0 {
		cfg.ShortOutAndBackMaxDistance = 5000
	}
	for i := 0; i < len(visits.Visits)-1; i++ {
		from, to := visits.Visits[i], visits.Visits[i+1]
		t := ExcursionLowConfidence
		if from.MatchedSite != "" && to.MatchedSite != "" {
			t = ExcursionBetweenKnownSites
		} else if from.MatchedSite != "" && to.MatchedSite == "" {
			t = ExcursionKnownToUnknown
		} else if from.MatchedSite == "" && to.MatchedSite != "" {
			t = ExcursionUnknownToKnown
		}
		d := gps.HaversineM(from.Latitude, from.Longitude, to.Latitude, to.Longitude)
		reasons := []ExcursionReason{ExcursionReasonVisitTransition}
		conf := ExcursionConfidenceMedium
		if from.Confidence == VisitConfidenceLow || to.Confidence == VisitConfidenceLow {
			reasons = append(reasons, ExcursionReasonLowVisitConf)
			conf = ExcursionConfidenceLow
			t = ExcursionLowConfidence
		}
		if hasVisitReason(from, VisitReasonGapInfluence) || hasVisitReason(to, VisitReasonGapInfluence) {
			reasons = append(reasons, ExcursionReasonGapInfluence)
			t = ExcursionGapAffected
		}
		if from.MatchedSite != "" && from.MatchedSite == to.MatchedSite && to.StartTime.Sub(from.EndTime) <= time.Duration(cfg.ShortOutAndBackMaxMinutes*float64(time.Minute)) && d <= cfg.ShortOutAndBackMaxDistance {
			t = ExcursionShortOutAndBack
			reasons = append(reasons, ExcursionReasonShortReturn)
			conf = ExcursionConfidenceHigh
		}
		if i >= 1 {
			prev := visits.Visits[i-1]
			if prev.MatchedSite != "" && prev.MatchedSite == to.MatchedSite && from.MatchedSite == "" && to.StartTime.Sub(prev.EndTime) <= time.Duration(cfg.ShortOutAndBackMaxMinutes*float64(time.Minute)) {
				t = ExcursionShortOutAndBack
				reasons = append(reasons, ExcursionReasonShortReturn)
				conf = ExcursionConfidenceHigh
			}
		}
		if from.Type == VisitSupplier || to.Type == VisitSupplier {
			t = ExcursionSupplierPickup
		}
		out.Excursions = append(out.Excursions, Excursion{FromVisitIndex: i, ToVisitIndex: i + 1, StartTime: from.EndTime, EndTime: to.StartTime, Duration: to.StartTime.Sub(from.EndTime), DistanceM: d, PointIndexes: append(append([]int(nil), from.PointIndexes...), to.PointIndexes...), Type: t, Reasons: reasons, Confidence: conf})
	}
	return out
}
func hasVisitReason(v Visit, reason VisitReason) bool {
	for _, r := range v.Reasons {
		if r == reason {
			return true
		}
	}
	return false
}
