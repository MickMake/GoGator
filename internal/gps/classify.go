package gps

import "gogator/internal/config"

func Classify(points []RawPoint, cfg config.Config) []RawPoint {
	ignitionMode := effectiveIgnitionMode(points, cfg)
	for i := range points {
		p := &points[i]
		io24, hasIO24 := Num(*p, "io24")
		io251, hasIO251 := Num(*p, "io251")
		ignition, hasIgnition := Num(*p, "io1")
		pdop, hasPDOP := Num(*p, "pdop")
		accel, hasAccel := AccelMagnitude(p.ParamNums)

		p.PDOPQuality = PDOPQuality(pdop, cfg.Trip.IdealPDOPThreshold, cfg.Trip.PoorPDOPThreshold, hasPDOP)

		if ignitionMode == "wired" && hasIgnition {
			if ignition == 0 {
				p.Stationary = true
				p.Flags = append(p.Flags, "ignition_off_stationary")
			}
			if ignition == 1 {
				p.Flags = append(p.Flags, "ignition_on")
			}
		}

		if hasIO24 && io24 == 1 {
			p.Moving = true
			p.Flags = append(p.Flags, "io24_moving")
		}
		if p.SpeedKPH >= cfg.Trip.MovingSpeedKPH {
			p.Moving = true
			p.Flags = append(p.Flags, "speed_moving")
		}
		if cfg.Trip.AccelerometerNonzeroSupportsMotion && hasAccel && accel > 0 {
			p.Moving = true
			p.Flags = append(p.Flags, "accel_nonzero")
		}

		if ignitionMode == "wired" && hasIgnition && ignition == 0 {
			p.Moving = false
			p.Stationary = true
		}

		if hasIO24 && io24 == 0 {
			p.Stationary = true
			p.Flags = append(p.Flags, "io24_stationary")
		}
		if p.SpeedKPH < cfg.Trip.MovingSpeedKPH && hasIO251 && io251 == 1 {
			p.Stationary = true
			p.Flags = append(p.Flags, "idling")
		}
		if hasPDOP && pdop > cfg.Trip.PoorPDOPThreshold {
			p.Flags = append(p.Flags, "poor_pdop")
		}
		if v, ok := Num(*p, "io247"); ok && v == 1 {
			p.Flags = append(p.Flags, "crash_detected")
		}
		if v, ok := Num(*p, "io253"); ok && v != 0 {
			p.Flags = append(p.Flags, "driver_behaviour")
		}
		if v, ok := Num(*p, "io3"); ok && v == 1 {
			p.Flags = append(p.Flags, "panic")
		}
		if v, ok := Num(*p, "io252"); ok && v == 1 {
			p.Flags = append(p.Flags, "power_cut")
		}
	}
	return points
}

func effectiveIgnitionMode(points []RawPoint, cfg config.Config) string {
	mode := cfg.Trip.IgnitionAnalysis
	if mode == "" {
		mode = "auto"
	}
	if mode == "wired" || mode == "unwired" {
		return mode
	}
	seenOn := false
	seenOff := false
	for _, p := range points {
		v, ok := Num(p, "io1")
		if !ok {
			continue
		}
		if v == 1 {
			seenOn = true
		}
		if v == 0 {
			seenOff = true
		}
		if seenOn && seenOff {
			return "wired"
		}
	}
	return "unwired"
}
