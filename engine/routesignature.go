package engine

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"gogator/engine/mapmatch"
	"gogator/internal/gps"
)

type RouteSignatureSource string

const (
	RouteSignatureSourceMatched RouteSignatureSource = "mapmatch"
	RouteSignatureSourceRaw     RouteSignatureSource = "candidate_trace"
)

type RouteSignatureWarning string

const (
	RouteSignatureWarningInsufficientPoints RouteSignatureWarning = "insufficient_points"
	RouteSignatureWarningNoSourcePoints     RouteSignatureWarning = "no_source_points"
)

type RouteSignatureCell struct{ ID string }

type RouteSignature struct {
	CandidateTripID  int
	Signature        string
	SignatureHash    string
	Source           RouteSignatureSource
	PointCount       int
	CellCount        int
	SourcePointStart int
	SourcePointEnd   int
	Cells            []RouteSignatureCell
	Warnings         []RouteSignatureWarning
}

func buildRouteSignatures(points []gps.RawPoint, trips []CandidateTrip, mm []CandidateTripMapMatchDiagnostic, enabled bool, resolution int) []RouteSignature {
	if !enabled || len(trips) == 0 {
		return nil
	}
	if resolution <= 0 {
		resolution = 7
	}
	out := make([]RouteSignature, 0, len(trips))
	for i, t := range trips {
		r := RouteSignature{CandidateTripID: i + 1, SourcePointStart: t.SourcePointStart, SourcePointEnd: t.SourcePointEnd}
		trace, src := signatureTrace(points, t, mapMatchShapeForTrip(mm, i+1))
		r.Source = src
		r.PointCount = len(trace)
		if len(trace) < 2 {
			r.Warnings = append(r.Warnings, RouteSignatureWarningInsufficientPoints)
			out = append(out, r)
			continue
		}
		for _, p := range trace {
			id := gridCellID(p.Lat, p.Lng, resolution)
			if id == "" {
				continue
			}
			if len(r.Cells) > 0 && r.Cells[len(r.Cells)-1].ID == id {
				continue
			}
			r.Cells = append(r.Cells, RouteSignatureCell{ID: id})
		}
		r.CellCount = len(r.Cells)
		if r.CellCount == 0 {
			r.Warnings = append(r.Warnings, RouteSignatureWarningNoSourcePoints)
			out = append(out, r)
			continue
		}
		parts := make([]string, 0, len(r.Cells))
		for _, c := range r.Cells {
			parts = append(parts, c.ID)
		}
		r.Signature = strings.Join(parts, ">")
		h := sha1.Sum([]byte(r.Signature))
		r.SignatureHash = hex.EncodeToString(h[:8])
		out = append(out, r)
	}
	return out
}

func signatureTrace(points []gps.RawPoint, t CandidateTrip, matched []mapmatch.MatchedShapePoint) ([]mapmatch.MatchPoint, RouteSignatureSource) {
	if len(matched) >= 2 {
		trace := make([]mapmatch.MatchPoint, 0, len(matched))
		for _, p := range matched {
			trace = append(trace, mapmatch.MatchPoint{Lat: p.Lat, Lng: p.Lng})
		}
		return trace, RouteSignatureSourceMatched
	}
	return candidateTripPoints(points, t), RouteSignatureSourceRaw
}
func mapMatchShapeForTrip(mm []CandidateTripMapMatchDiagnostic, id int) []mapmatch.MatchedShapePoint {
	for _, d := range mm {
		if d.CandidateTripID == id {
			return d.MatchedShape
		}
	}
	return nil
}
func gridCellID(lat, lng float64, resolution int) string {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return ""
	}
	if resolution <= 0 {
		resolution = 7
	}
	scale := math.Pow(2, float64(resolution))
	latB := int(math.Floor((lat + 90.0) * scale))
	lngB := int(math.Floor((lng + 180.0) * scale))
	return fmt.Sprintf("grid:%d:%d:%d", resolution, latB, lngB)
}
