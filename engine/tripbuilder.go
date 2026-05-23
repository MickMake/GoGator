package engine

import "time"

type CandidateTripType string
type CandidateTripReason string
type CandidateTripConfidence string
type CandidateTripBoundary string

const (
	CandidateTripSiteToSite       CandidateTripType = "SiteToSite"
	CandidateTripSiteToUnknown    CandidateTripType = "SiteToUnknown"
	CandidateTripUnknownToSite    CandidateTripType = "UnknownToSite"
	CandidateTripUnknownToUnknown CandidateTripType = "UnknownToUnknown"
	CandidateTripOutAndBack       CandidateTripType = "OutAndBack"
	CandidateTripGapAffected      CandidateTripType = "GapAffected"
	CandidateTripLowConfidence    CandidateTripType = "LowConfidence"
	CandidateTripNoiseAffected    CandidateTripType = "NoiseAffected"
)
const (
	CandidateReasonVisitTransition     CandidateTripReason = "visit_transition"
	CandidateReasonGapInfluence        CandidateTripReason = "gap_influence"
	CandidateReasonBoundaryNoise       CandidateTripReason = "boundary_noise"
	CandidateReasonDurationImplausible CandidateTripReason = "duration_implausible"
)
const (
	CandidateConfidenceLow    CandidateTripConfidence = "Low"
	CandidateConfidenceMedium CandidateTripConfidence = "Medium"
	CandidateConfidenceHigh   CandidateTripConfidence = "High"
)
const (
	BoundaryLow    CandidateTripBoundary = "Low"
	BoundaryMedium CandidateTripBoundary = "Medium"
	BoundaryHigh   CandidateTripBoundary = "High"
)

type CandidateTrip struct {
	StartTime             time.Time
	EndTime               time.Time
	Duration              time.Duration
	OriginVisitIndex      int
	DestinationVisitIndex int
	OriginLabel           string
	DestinationLabel      string
	SourcePointStart      int
	SourcePointEnd        int
	ApproxDistanceM       float64
	Type                  CandidateTripType
	Confidence            CandidateTripConfidence
	OriginBoundary        CandidateTripBoundary
	DestinationBoundary   CandidateTripBoundary
	MovementBoundary      CandidateTripBoundary
	GPSQualityBoundary    CandidateTripBoundary
	GapNoiseBoundary      CandidateTripBoundary
	SiteMatchBoundary     CandidateTripBoundary
	DurationBoundary      CandidateTripBoundary
	Reasons               []CandidateTripReason
	Warnings              []string
}

type CandidateTripEvidence struct{ Trips []CandidateTrip }

type TripBuilderConfig struct {
	Enabled                bool
	PassiveOnly            bool
	CompareLegacy          bool
	MinTripDurationMinutes float64
	MaxGapMinutes          float64
	LowConfidenceThreshold float64
}

func detectCandidateTrips(visits VisitEvidence, excursions ExcursionEvidence, cfg TripBuilderConfig) CandidateTripEvidence {
	out := CandidateTripEvidence{}
	if !cfg.Enabled || len(visits.Visits) < 2 {
		return out
	}
	if cfg.MinTripDurationMinutes <= 0 {
		cfg.MinTripDurationMinutes = 1
	}
	for i := 0; i < len(visits.Visits)-1; i++ {
		from, to := visits.Visits[i], visits.Visits[i+1]
		dur := to.StartTime.Sub(from.EndTime)
		ct := CandidateTrip{StartTime: from.EndTime, EndTime: to.StartTime, Duration: dur, OriginVisitIndex: i, DestinationVisitIndex: i + 1, OriginLabel: from.MatchedSite, DestinationLabel: to.MatchedSite, Confidence: CandidateConfidenceMedium, OriginBoundary: BoundaryMedium, DestinationBoundary: BoundaryMedium, MovementBoundary: BoundaryMedium, GPSQualityBoundary: BoundaryMedium, GapNoiseBoundary: BoundaryMedium, SiteMatchBoundary: BoundaryMedium, DurationBoundary: BoundaryMedium, Reasons: []CandidateTripReason{CandidateReasonVisitTransition}}
		if len(from.PointIndexes) > 0 {
			ct.SourcePointStart = from.PointIndexes[0]
		}
		if len(to.PointIndexes) > 0 {
			ct.SourcePointEnd = to.PointIndexes[len(to.PointIndexes)-1]
		}
		if from.MatchedSite != "" && to.MatchedSite != "" {
			ct.Type = CandidateTripSiteToSite
		} else if from.MatchedSite != "" {
			ct.Type = CandidateTripSiteToUnknown
		} else if to.MatchedSite != "" {
			ct.Type = CandidateTripUnknownToSite
		} else {
			ct.Type = CandidateTripUnknownToUnknown
		}
		for _, ex := range excursions.Excursions {
			if ex.FromVisitIndex == i && ex.ToVisitIndex == i+1 {
				ct.ApproxDistanceM = ex.DistanceM
				if ex.Type == ExcursionShortOutAndBack {
					ct.Type = CandidateTripOutAndBack
				}
				if ex.Type == ExcursionGapAffected {
					ct.Type = CandidateTripGapAffected
					ct.Reasons = append(ct.Reasons, CandidateReasonGapInfluence)
					ct.GapNoiseBoundary = BoundaryLow
				}
				break
			}
		}
		if from.Confidence == VisitConfidenceLow || to.Confidence == VisitConfidenceLow {
			ct.Confidence = CandidateConfidenceLow
			ct.Type = CandidateTripLowConfidence
		}
		if hasVisitReason(from, VisitReasonGapInfluence) || hasVisitReason(to, VisitReasonGapInfluence) {
			ct.Type = CandidateTripGapAffected
			ct.Reasons = append(ct.Reasons, CandidateReasonGapInfluence)
			ct.GapNoiseBoundary = BoundaryLow
		}
		if hasVisitReason(from, VisitReasonFromStayType) || hasVisitReason(to, VisitReasonFromStayType) {
			ct.Type = CandidateTripNoiseAffected
			ct.Reasons = append(ct.Reasons, CandidateReasonBoundaryNoise)
			ct.OriginBoundary = BoundaryLow
			ct.DestinationBoundary = BoundaryLow
		}
		if dur < time.Duration(cfg.MinTripDurationMinutes*float64(time.Minute)) {
			ct.DurationBoundary = BoundaryLow
			ct.Warnings = append(ct.Warnings, "short_duration")
			ct.Reasons = append(ct.Reasons, CandidateReasonDurationImplausible)
		}
		if ct.OriginLabel == "" || ct.DestinationLabel == "" {
			ct.SiteMatchBoundary = BoundaryLow
		}
		out.Trips = append(out.Trips, ct)
	}
	return out
}
