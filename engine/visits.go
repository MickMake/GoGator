package engine

import (
	"strings"
	"time"

	"gogator/internal/sites"
)

type VisitConfig struct {
	Enabled                 bool
	MinVisitDurationMinutes float64
}

type VisitType string
type VisitReason string
type VisitConfidence string

const (
	VisitKnownSite VisitType = "KnownSiteVisit"
	VisitUnknown   VisitType = "UnknownVisit"
	VisitHome      VisitType = "HomeVisit"
	VisitSupplier  VisitType = "SupplierVisit"
	VisitPause     VisitType = "PauseVisit"
	VisitTraffic   VisitType = "TrafficPause"
	VisitNoise     VisitType = "NoiseVisit"
)

const (
	VisitReasonFromSiteMatch  VisitReason = "from_site_match"
	VisitReasonFromStayType   VisitReason = "from_stay_type"
	VisitReasonHomeSite       VisitReason = "home_site"
	VisitReasonSupplierSite   VisitReason = "supplier_site"
	VisitReasonShortDuration  VisitReason = "short_duration"
	VisitReasonLowStayQuality VisitReason = "low_stay_quality"
	VisitReasonGapInfluence   VisitReason = "gap_influence"
)

const (
	VisitConfidenceLow    VisitConfidence = "Low"
	VisitConfidenceMedium VisitConfidence = "Medium"
	VisitConfidenceHigh   VisitConfidence = "High"
)

type Visit struct {
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Latitude     float64
	Longitude    float64
	MatchedSite  string
	StayIndex    int
	StayType     StayType
	Type         VisitType
	Reasons      []VisitReason
	Confidence   VisitConfidence
	PointIndexes []int
}

type VisitSegment struct {
	VisitIndexes []int
}

type VisitEvidence struct {
	Visits   []Visit
	Segments []VisitSegment
}

func detectVisits(stays StayEvidence, siteList []sites.Site, cfg VisitConfig) VisitEvidence {
	if !cfg.Enabled {
		return VisitEvidence{}
	}
	if cfg.MinVisitDurationMinutes <= 0 {
		cfg.MinVisitDurationMinutes = 5
	}

	out := VisitEvidence{Visits: make([]Visit, 0, len(stays.Stays))}
	for i, st := range stays.Stays {
		v := Visit{StartTime: st.StartTime, EndTime: st.EndTime, Duration: st.Duration, Latitude: st.Latitude, Longitude: st.Longitude, MatchedSite: st.MatchedSite, StayIndex: i, StayType: st.Type, Type: VisitUnknown, Confidence: visitConfidenceFromStay(st.Confidence), PointIndexes: append([]int(nil), st.PointIndexes...)}
		if st.MatchedSite != "" {
			v.Type = VisitKnownSite
			v.Reasons = append(v.Reasons, VisitReasonFromSiteMatch)
		}
		switch st.Type {
		case StayTypePause:
			v.Type = VisitPause
			v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonFromStayType)
		case StayTypeTraffic:
			v.Type = VisitTraffic
			v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonFromStayType)
		case StayTypeNoiseCluster:
			v.Type = VisitNoise
			v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonFromStayType)
		case StayTypeGapInferredStop:
			v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonGapInfluence)
		}
		if st.Duration < time.Duration(cfg.MinVisitDurationMinutes*float64(time.Minute)) {
			v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonShortDuration)
		}
		if st.Confidence == StayConfidenceLow {
			v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonLowStayQuality)
		}
		if site, ok := findSiteByName(siteList, st.MatchedSite); ok {
			t := strings.ToLower(strings.TrimSpace(site.SiteType))
			if strings.Contains(t, "home") {
				v.Type = VisitHome
				v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonHomeSite)
			}
			if strings.Contains(t, "supplier") || strings.Contains(t, "vendor") || strings.Contains(t, "pickup") {
				v.Type = VisitSupplier
				v.Reasons = appendUniqueVisitReason(v.Reasons, VisitReasonSupplierSite)
			}
		}
		out.Visits = append(out.Visits, v)
		out.Segments = append(out.Segments, VisitSegment{VisitIndexes: []int{i}})
	}
	return out
}

func findSiteByName(siteList []sites.Site, name string) (sites.Site, bool) {
	for _, s := range siteList {
		if s.Name == name {
			return s, true
		}
	}
	return sites.Site{}, false
}
func visitConfidenceFromStay(conf StayConfidence) VisitConfidence {
	switch conf {
	case StayConfidenceHigh:
		return VisitConfidenceHigh
	case StayConfidenceLow:
		return VisitConfidenceLow
	default:
		return VisitConfidenceMedium
	}
}
func appendUniqueVisitReason(in []VisitReason, reason VisitReason) []VisitReason {
	for _, r := range in {
		if r == reason {
			return in
		}
	}
	return append(in, reason)
}
