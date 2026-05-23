package engine

import "time"

type PointEvidence struct {
	Index       int
	SourceFile  string
	RawRow      int
	Time        time.Time
	Coordinates CoordinateEvidence
	Signals     SignalEvidence
	Quality     QualityScore
}

type CoordinateEvidence struct {
	Lat     float64
	Lng     float64
	Missing bool
	Invalid bool
}

type SignalEvidence struct {
	IO24   *float64
	IO251  *float64
	IO14   *float64
	PDOP   *float64
	GPSLev *float64
	GSMLev *float64
	G0     *float64
	G1     *float64
	G2     *float64
	IO247  *float64
	IO253  *float64
	IO303  *float64
}

type EvidenceSet struct {
	Points []PointEvidence
}
