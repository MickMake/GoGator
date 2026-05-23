package engine

import "gogator/internal/gps"

func extractSignalEvidence(p gps.RawPoint) SignalEvidence {
	return SignalEvidence{
		IO24:   signalPtr(p, "io24"),
		IO251:  signalPtr(p, "io251"),
		IO14:   signalPtr(p, "io14"),
		PDOP:   signalPtr(p, "pdop"),
		GPSLev: signalPtr(p, "gpslev"),
		GSMLev: signalPtr(p, "gsmlev"),
		G0:     signalPtr(p, "g0"),
		G1:     signalPtr(p, "g1"),
		G2:     signalPtr(p, "g2"),
		IO247:  signalPtr(p, "io247"),
		IO253:  signalPtr(p, "io253"),
		IO303:  signalPtr(p, "io303"),
	}
}

func signalPtr(p gps.RawPoint, key string) *float64 {
	if p.ParamNums == nil {
		return nil
	}
	v, ok := p.ParamNums[key]
	if !ok {
		return nil
	}
	vv := v
	return &vv
}
