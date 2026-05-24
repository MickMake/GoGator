package engine

import "time"

type MotionState string

type MotionReason string

const (
	MotionUnknown    MotionState = "Unknown"
	MotionMoving     MotionState = "Moving"
	MotionStationary MotionState = "Stationary"
	MotionGap        MotionState = "Gap"
	MotionNoise      MotionState = "Noise"
)

const (
	MotionReasonInsufficientEvidence MotionReason = "insufficient_evidence"
	MotionReasonSpeedMoving          MotionReason = "speed_moving"
	MotionReasonSpeedStationary      MotionReason = "speed_stationary"
	MotionReasonSignalMoving         MotionReason = "signal_moving"
	MotionReasonSignalStationary     MotionReason = "signal_stationary"
	MotionReasonQualityPoor          MotionReason = "quality_poor"
	MotionReasonInvalidPoint         MotionReason = "invalid_point"
	MotionReasonTimeGap              MotionReason = "time_gap"
	MotionReasonDuplicateTime        MotionReason = "duplicate_or_non_increasing_time"
)

type MotionConfig struct {
	Enabled                     bool
	StationarySpeedThresholdKPH float64
	MovingSpeedThresholdKPH     float64
	GapThresholdMinutes         float64
	MinConsecutiveSamples       int
}

type MotionSample struct {
	Index    int
	Time     time.Time
	State    MotionState
	SpeedKPH float64
	Reason   MotionReason
	Quality  QualityBand
}

type MotionSegment struct {
	State      MotionState
	StartIndex int
	EndIndex   int
	StartTime  time.Time
	EndTime    time.Time
	Reasons    []MotionReason
}

type MotionEvidence struct {
	Samples  []MotionSample
	Segments []MotionSegment
}

func classifyMotion(points []PointEvidence, cfg MotionConfig) MotionEvidence {
	out := MotionEvidence{Samples: make([]MotionSample, 0, len(points))}
	if len(points) == 0 {
		return out
	}
	if cfg.MinConsecutiveSamples < 1 {
		cfg.MinConsecutiveSamples = 1
	}
	pending := MotionUnknown
	pendingCount := 0
	active := MotionUnknown

	for i, p := range points {
		state, reason := classifyInstantMotion(i, points, cfg)
		if state == MotionGap || state == MotionNoise {
			pending = MotionUnknown
			pendingCount = 0
		} else if state == MotionUnknown {
			pending = MotionUnknown
			pendingCount = 0
		} else {
			if active == MotionUnknown {
				active = state
				pending = MotionUnknown
				pendingCount = 0
			} else if active != state {
				if pending == state {
					pendingCount++
				} else {
					pending = state
					pendingCount = 1
				}
				if pendingCount >= cfg.MinConsecutiveSamples {
					active = state
					pending = MotionUnknown
					pendingCount = 0
				}
			} else {
				pending = MotionUnknown
				pendingCount = 0
			}
		}

		effective := active
		if state == MotionGap || state == MotionNoise {
			effective = state
		}
		if active != MotionGap && active != MotionNoise && state == MotionUnknown {
			effective = active
		}
		if active == MotionUnknown {
			effective = state
		}

		speedKPH := 0.0
		if p.Signals.Speed != nil {
			speedKPH = *p.Signals.Speed
		}
		out.Samples = append(out.Samples, MotionSample{Index: p.Index, Time: p.Time, State: effective, SpeedKPH: speedKPH, Reason: reason, Quality: p.Quality.Band})
	}
	out.Segments = buildMotionSegments(out.Samples)
	return out
}

func classifyInstantMotion(i int, points []PointEvidence, cfg MotionConfig) (MotionState, MotionReason) {
	p := points[i]
	if p.Coordinates.Missing || p.Coordinates.Invalid || p.Quality.Band == QualityInvalid {
		return MotionNoise, MotionReasonInvalidPoint
	}
	if i > 0 {
		gap := p.Time.Sub(points[i-1].Time)
		if gap <= 0 {
			return MotionNoise, MotionReasonDuplicateTime
		}
		if gap >= time.Duration(cfg.GapThresholdMinutes*float64(time.Minute)) {
			return MotionGap, MotionReasonTimeGap
		}
	}
	if p.Quality.Band == QualityPoor {
		return MotionUnknown, MotionReasonQualityPoor
	}

	movingEvidence := 0
	stationaryEvidence := 0
	if speed := p.Signals.Speed; speed != nil {
		if *speed >= cfg.MovingSpeedThresholdKPH {
			movingEvidence += 2
		}
		if *speed <= cfg.StationarySpeedThresholdKPH {
			stationaryEvidence += 2
		}
	}
	if p.Signals.IO24 != nil {
		if *p.Signals.IO24 >= 1 {
			movingEvidence += 2
		} else if *p.Signals.IO24 == 0 {
			stationaryEvidence += 2
		}
	}
	if p.Signals.IO251 != nil && *p.Signals.IO251 >= 1 {
		stationaryEvidence++
	}

	switch {
	case movingEvidence >= stationaryEvidence+1 && movingEvidence >= 2:
		if p.Signals.Speed != nil && *p.Signals.Speed >= cfg.MovingSpeedThresholdKPH {
			return MotionMoving, MotionReasonSpeedMoving
		}
		return MotionMoving, MotionReasonSignalMoving
	case stationaryEvidence >= movingEvidence+1 && stationaryEvidence >= 2:
		if p.Signals.Speed != nil && *p.Signals.Speed <= cfg.StationarySpeedThresholdKPH {
			return MotionStationary, MotionReasonSpeedStationary
		}
		return MotionStationary, MotionReasonSignalStationary
	default:
		return MotionUnknown, MotionReasonInsufficientEvidence
	}
}

func buildMotionSegments(samples []MotionSample) []MotionSegment {
	if len(samples) == 0 {
		return nil
	}
	segments := []MotionSegment{}
	cur := MotionSegment{State: samples[0].State, StartIndex: samples[0].Index, EndIndex: samples[0].Index, StartTime: samples[0].Time, EndTime: samples[0].Time, Reasons: []MotionReason{samples[0].Reason}}
	for i := 1; i < len(samples); i++ {
		s := samples[i]
		if s.State != cur.State {
			segments = append(segments, cur)
			cur = MotionSegment{State: s.State, StartIndex: s.Index, EndIndex: s.Index, StartTime: s.Time, EndTime: s.Time, Reasons: []MotionReason{s.Reason}}
			continue
		}
		cur.EndIndex = s.Index
		cur.EndTime = s.Time
		if len(cur.Reasons) == 0 || cur.Reasons[len(cur.Reasons)-1] != s.Reason {
			cur.Reasons = append(cur.Reasons, s.Reason)
		}
	}
	segments = append(segments, cur)
	return segments
}
