package gps

import "gogator/internal/config"

func Classify(points []RawPoint, cfg config.Config) []RawPoint {
	for i := range points {
		p := &points[i]
		io24, hasIO24 := Num(*p, "io24")
		io251, hasIO251 := Num(*p, "io251")
		pdop, hasPDOP := Num(*p, "pdop")
		accel, hasAccel := AccelMagnitude(p.ParamNums)

		p.PDOPQuality = PDOPQuality(pdop, cfg.Trip.IdealPDOPThreshold, cfg.Trip.PoorPDOPThreshold, hasPDOP)
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
