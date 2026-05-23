package engine

import (
	"math"
	"time"

	"gogator/internal/gps"
)

type QualityBand string

const (
	QualityGood    QualityBand = "Good"
	QualityUsable  QualityBand = "Usable"
	QualityPoor    QualityBand = "Poor"
	QualityInvalid QualityBand = "Invalid"
	QualityUnknown QualityBand = "Unknown"
)

type QualityReason string

const (
	ReasonMissingCoordinates  QualityReason = "missing_coordinates"
	ReasonInvalidCoordinates  QualityReason = "invalid_coordinates"
	ReasonDuplicateTimestamp  QualityReason = "duplicate_or_non_increasing_timestamp"
	ReasonLargeTimeGap        QualityReason = "large_time_gap"
	ReasonPoorPDOP            QualityReason = "poor_pdop"
	ReasonLowGPSLevel         QualityReason = "low_gps_level"
	ReasonWeakGSMLevel        QualityReason = "weak_gsm_level"
	ReasonImprobableSpeedJump QualityReason = "improbable_speed_jump"
	ReasonGoodSignalMix       QualityReason = "good_signal_mix"
	ReasonSparseEvidence      QualityReason = "sparse_signal_evidence"
)

type QualityScore struct {
	Band    QualityBand
	Score   int
	Reasons []QualityReason
}

func buildEvidence(points []gps.RawPoint, enableQuality bool) EvidenceSet {
	set := EvidenceSet{Points: make([]PointEvidence, 0, len(points))}
	for i, p := range points {
		ev := PointEvidence{
			Index:      i,
			SourceFile: p.SourceFile,
			RawRow:     p.RawRow,
			Time:       p.Time,
			Coordinates: CoordinateEvidence{Lat: p.Lat, Lng: p.Lng,
				Missing: p.Lat == 0 && p.Lng == 0,
				Invalid: !validCoordinate(p.Lat, p.Lng),
			},
			Signals: extractSignalEvidence(p),
		}
		s := p.SpeedKPH
		ev.Signals.Speed = &s
		if enableQuality {
			ev.Quality = scorePointQuality(i, points, ev)
		} else {
			ev.Quality = QualityScore{Band: QualityUnknown}
		}
		set.Points = append(set.Points, ev)
	}
	return set
}

func scorePointQuality(i int, points []gps.RawPoint, ev PointEvidence) QualityScore {
	score := 50
	reasons := []QualityReason{}
	if ev.Coordinates.Missing {
		return QualityScore{Band: QualityInvalid, Score: 0, Reasons: []QualityReason{ReasonMissingCoordinates}}
	}
	if ev.Coordinates.Invalid {
		return QualityScore{Band: QualityInvalid, Score: 0, Reasons: []QualityReason{ReasonInvalidCoordinates}}
	}
	if i > 0 {
		prev := points[i-1]
		if !ev.Time.After(prev.Time) {
			score -= 30
			reasons = append(reasons, ReasonDuplicateTimestamp)
		}
		gap := ev.Time.Sub(prev.Time)
		if gap > 2*time.Hour {
			score -= 10
			reasons = append(reasons, ReasonLargeTimeGap)
		}
		if p := points[i]; p.ImpliedSpeedKPH > 180 {
			score -= 25
			reasons = append(reasons, ReasonImprobableSpeedJump)
		}
	}
	if ev.Signals.PDOP != nil {
		if *ev.Signals.PDOP > 5 {
			score -= 20
			reasons = append(reasons, ReasonPoorPDOP)
		} else if *ev.Signals.PDOP <= 2 {
			score += 10
		}
	}
	if ev.Signals.GPSLev != nil {
		if *ev.Signals.GPSLev <= 1 {
			score -= 20
			reasons = append(reasons, ReasonLowGPSLevel)
		} else if *ev.Signals.GPSLev >= 4 {
			score += 8
		}
	}
	if ev.Signals.GSMLev != nil {
		if *ev.Signals.GSMLev <= 5 {
			score -= 8
			reasons = append(reasons, ReasonWeakGSMLevel)
		} else if *ev.Signals.GSMLev >= 15 {
			score += 5
		}
	}
	if ev.Signals.IO24 != nil && *ev.Signals.IO24 == 1 && points[i].SpeedKPH > 5 {
		score += 6
	}
	if len(reasons) == 0 {
		reasons = append(reasons, ReasonGoodSignalMix)
	}
	if signalCount(ev.Signals) < 2 {
		reasons = append(reasons, ReasonSparseEvidence)
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return QualityScore{Band: classifyBand(score), Score: score, Reasons: reasons}
}

func signalCount(s SignalEvidence) int {
	count := 0
	for _, p := range []*float64{s.IO24, s.IO251, s.IO14, s.PDOP, s.GPSLev, s.GSMLev, s.G0, s.G1, s.G2, s.IO247, s.IO253, s.IO303} {
		if p != nil {
			count++
		}
	}
	return count
}

func classifyBand(score int) QualityBand {
	switch {
	case score >= 75:
		return QualityGood
	case score >= 60:
		return QualityUsable
	case score > 0:
		return QualityPoor
	default:
		return QualityInvalid
	}
}

func validCoordinate(lat, lng float64) bool {
	if math.IsNaN(lat) || math.IsNaN(lng) {
		return false
	}
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
