package engine

import (
	"math"
	"sort"
	"time"

	"gogator/internal/gps"
	"gogator/internal/sites"
)

type StayType string

type StayReason string

type StayConfidence string

const (
	StayTypeSiteStop        StayType = "SiteStop"
	StayTypeUnknownStop     StayType = "UnknownStop"
	StayTypePause           StayType = "Pause"
	StayTypeTraffic         StayType = "Traffic"
	StayTypePickupCandidate StayType = "PickupCandidate"
	StayTypeGapInferredStop StayType = "GapInferredStop"
	StayTypeNoiseCluster    StayType = "NoiseCluster"
)

const (
	StayReasonStationaryMotion StayReason = "stationary_motion"
	StayReasonLowSpeed         StayReason = "low_speed"
	StayReasonTightRadius      StayReason = "tight_radius"
	StayReasonScatteredRadius  StayReason = "scattered_radius"
	StayReasonShortDuration    StayReason = "short_duration"
	StayReasonPoorQuality      StayReason = "poor_quality"
	StayReasonTimeGap          StayReason = "time_gap"
	StayReasonNearKnownSite    StayReason = "near_known_site"
)

const (
	StayConfidenceLow    StayConfidence = "Low"
	StayConfidenceMedium StayConfidence = "Medium"
	StayConfidenceHigh   StayConfidence = "High"
)

type StayConfig struct {
	Enabled                bool
	MinDurationMinutes     float64
	MaxRadiusMeters        float64
	MinPoints              int
	SiteMatchRadiusMeters  float64
	GapInferredStopEnabled bool
}

type Stay struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	Latitude  float64
	Longitude float64
	RadiusM   float64

	PointCount    int
	Confidence    StayConfidence
	Type          StayType
	Reasons       []StayReason
	PointIndexes  []int
	MatchedSite   string
	MatchedRadius float64
}

type StayEvidence struct{ Stays []Stay }

func detectStays(points []PointEvidence, motion MotionEvidence, siteList []sites.Site, cfg StayConfig) StayEvidence {
	out := StayEvidence{}
	if !cfg.Enabled || len(points) == 0 {
		return out
	}
	if cfg.MinPoints < 2 {
		cfg.MinPoints = 2
	}
	if cfg.MinDurationMinutes <= 0 {
		cfg.MinDurationMinutes = 5
	}
	if cfg.MaxRadiusMeters <= 0 {
		cfg.MaxRadiusMeters = 120
	}
	if cfg.SiteMatchRadiusMeters <= 0 {
		cfg.SiteMatchRadiusMeters = 100
	}
	for _, seg := range motion.Segments {
		if seg.State != MotionStationary && seg.State != MotionNoise {
			continue
		}
		cluster := sampleWindow(points, seg.StartIndex, seg.EndIndex)
		if len(cluster) < cfg.MinPoints {
			continue
		}
		stay := summarizeStay(cluster, cfg)
		if len(siteList) > 0 {
			if name, dist, ok := matchSite(stay.Latitude, stay.Longitude, siteList, cfg.SiteMatchRadiusMeters); ok {
				stay.Type = StayTypeSiteStop
				stay.MatchedSite = name
				stay.MatchedRadius = dist
				stay.Reasons = appendUniqueStayReason(stay.Reasons, StayReasonNearKnownSite)
			}
		}
		out.Stays = append(out.Stays, stay)
	}
	if cfg.GapInferredStopEnabled {
		out.Stays = append(out.Stays, inferGapStays(points, motion, cfg)...)
	}
	sort.SliceStable(out.Stays, func(i, j int) bool { return out.Stays[i].StartTime.Before(out.Stays[j].StartTime) })
	return out
}

func sampleWindow(points []PointEvidence, startIdx, endIdx int) []PointEvidence {
	out := make([]PointEvidence, 0, endIdx-startIdx+1)
	for _, p := range points {
		if p.Index >= startIdx && p.Index <= endIdx {
			out = append(out, p)
		}
	}
	return out
}

func summarizeStay(cluster []PointEvidence, cfg StayConfig) Stay {
	start := cluster[0].Time
	end := cluster[len(cluster)-1].Time
	duration := end.Sub(start)
	lat, lng := centroid(cluster)
	radius := clusterRadius(cluster, lat, lng)
	reasons := []StayReason{StayReasonStationaryMotion}
	stayType := StayTypeUnknownStop
	if duration < time.Duration(cfg.MinDurationMinutes*float64(time.Minute)) {
		reasons = append(reasons, StayReasonShortDuration)
		stayType = StayTypePause
	}
	if radius > cfg.MaxRadiusMeters {
		reasons = append(reasons, StayReasonScatteredRadius)
		stayType = StayTypeNoiseCluster
	} else {
		reasons = append(reasons, StayReasonTightRadius)
	}
	poor := 0
	idxs := make([]int, 0, len(cluster))
	for _, p := range cluster {
		idxs = append(idxs, p.Index)
		if p.Quality.Band == QualityPoor || p.Quality.Band == QualityInvalid {
			poor++
		}
		if p.Signals.Speed != nil && *p.Signals.Speed <= 2 {
			reasons = appendUniqueStayReason(reasons, StayReasonLowSpeed)
		}
	}
	if poor*2 >= len(cluster) {
		reasons = append(reasons, StayReasonPoorQuality)
	}
	conf := StayConfidenceMedium
	if stayType == StayTypePause || stayType == StayTypeNoiseCluster || poor*2 >= len(cluster) {
		conf = StayConfidenceLow
	} else if duration >= 2*time.Duration(cfg.MinDurationMinutes*float64(time.Minute)) && radius <= cfg.MaxRadiusMeters*0.6 {
		conf = StayConfidenceHigh
	}
	return Stay{StartTime: start, EndTime: end, Duration: duration, Latitude: lat, Longitude: lng, RadiusM: radius, PointCount: len(cluster), Confidence: conf, Type: stayType, Reasons: reasons, PointIndexes: idxs}
}

func inferGapStays(points []PointEvidence, motion MotionEvidence, cfg StayConfig) []Stay {
	var out []Stay
	for _, s := range motion.Samples {
		if s.State != MotionGap {
			continue
		}
		cur := byIndex(points, s.Index)
		prev := byIndex(points, s.Index-1)
		if cur == nil || prev == nil {
			continue
		}
		d := cur.Time.Sub(prev.Time)
		if d < time.Duration(cfg.MinDurationMinutes*float64(time.Minute)) {
			continue
		}
		out = append(out, Stay{StartTime: prev.Time, EndTime: cur.Time, Duration: d, Latitude: (prev.Coordinates.Lat + cur.Coordinates.Lat) / 2, Longitude: (prev.Coordinates.Lng + cur.Coordinates.Lng) / 2, RadiusM: gps.HaversineM(prev.Coordinates.Lat, prev.Coordinates.Lng, cur.Coordinates.Lat, cur.Coordinates.Lng) / 2, PointCount: 2, Confidence: StayConfidenceLow, Type: StayTypeGapInferredStop, Reasons: []StayReason{StayReasonTimeGap}, PointIndexes: []int{prev.Index, cur.Index}})
	}
	return out
}
func byIndex(points []PointEvidence, idx int) *PointEvidence {
	for i := range points {
		if points[i].Index == idx {
			return &points[i]
		}
	}
	return nil
}
func centroid(points []PointEvidence) (float64, float64) {
	var lat, lng float64
	var n float64
	for _, p := range points {
		if p.Coordinates.Missing || p.Coordinates.Invalid {
			continue
		}
		lat += p.Coordinates.Lat
		lng += p.Coordinates.Lng
		n++
	}
	if n == 0 {
		return 0, 0
	}
	return lat / n, lng / n
}
func clusterRadius(points []PointEvidence, lat, lng float64) float64 {
	m := 0.0
	for _, p := range points {
		d := gps.HaversineM(lat, lng, p.Coordinates.Lat, p.Coordinates.Lng)
		if d > m {
			m = d
		}
	}
	return m
}
func appendUniqueStayReason(in []StayReason, reason StayReason) []StayReason {
	for _, r := range in {
		if r == reason {
			return in
		}
	}
	return append(in, reason)
}
func matchSite(lat, lng float64, siteList []sites.Site, max float64) (string, float64, bool) {
	best := math.MaxFloat64
	name := ""
	for _, s := range siteList {
		d := gps.HaversineM(lat, lng, s.Lat, s.Lng)
		lim := max
		if s.RadiusM > 0 && s.RadiusM < lim {
			lim = s.RadiusM
		}
		if d <= lim && d < best {
			best = d
			name = s.Name
		}
	}
	if name == "" {
		return "", 0, false
	}
	return name, best, true
}
