package engine

import (
	"fmt"
	"strings"

	"gogator/engine/mapmatch"
	"gogator/internal/gps"
)

type CandidateTripMapMatchDiagnostic struct {
	CandidateTripID       int
	Matched               bool
	MatchedDistanceMeters float64
	MatchedDurationSecs   float64
	EdgeCount             int
	WarningCount          int
	WarningCodes          []string
	Error                 string
	Confidence            float64
}

func buildMapMatchDiagnostics(matcher mapmatch.MapMatcher, points []gps.RawPoint, trips []CandidateTrip, cfg ValhallaConfig) []CandidateTripMapMatchDiagnostic {
	if len(trips) == 0 {
		return nil
	}
	out := make([]CandidateTripMapMatchDiagnostic, 0, len(trips))
	for i, ct := range trips {
		d := CandidateTripMapMatchDiagnostic{CandidateTripID: i + 1}
		trace := candidateTripPoints(points, ct)
		if len(trace) < 2 {
			d.Error = "insufficient_points"
			out = append(out, d)
			continue
		}
		res, err := matcher.Match(mapmatch.MatchRequest{Points: trace, Endpoint: cfg.Endpoint, MaxPointsPerRequest: cfg.MaxPointsPerRequest})
		if err != nil {
			d.Error = err.Error()
			out = append(out, d)
			continue
		}
		d.MatchedDistanceMeters = res.DistanceM
		d.MatchedDurationSecs = res.DurationS
		d.EdgeCount = len(res.MatchedEdges)
		d.WarningCount = len(res.Warnings)
		d.Confidence = res.Confidence.Score
		d.WarningCodes = make([]string, 0, len(res.Warnings))
		for _, w := range res.Warnings {
			code := strings.TrimSpace(w.Code)
			if code != "" {
				d.WarningCodes = append(d.WarningCodes, code)
			}
		}
		d.Matched = d.Error == "" && (d.MatchedDistanceMeters > 0 || d.MatchedDurationSecs > 0 || d.EdgeCount > 0)
		out = append(out, d)
	}
	return out
}

func candidateTripPoints(points []gps.RawPoint, ct CandidateTrip) []mapmatch.MatchPoint {
	if ct.SourcePointStart < 0 || ct.SourcePointEnd < ct.SourcePointStart || ct.SourcePointEnd >= len(points) {
		return nil
	}
	trace := make([]mapmatch.MatchPoint, 0, ct.SourcePointEnd-ct.SourcePointStart+1)
	for idx := ct.SourcePointStart; idx <= ct.SourcePointEnd; idx++ {
		p := points[idx]
		if p.Lat < -90 || p.Lat > 90 || p.Lng < -180 || p.Lng > 180 {
			continue
		}
		trace = append(trace, mapmatch.MatchPoint{Lat: p.Lat, Lng: p.Lng, Time: p.Time})
	}
	return trace
}

func summarizeMapMatchWarnings(rows []CandidateTripMapMatchDiagnostic) []string {
	warnings := []string{}
	for _, r := range rows {
		if r.Error != "" && r.Error != "insufficient_points" {
			warnings = append(warnings, fmt.Sprintf("candidate_trip_id=%d mapmatch_error=%s", r.CandidateTripID, r.Error))
		}
	}
	return warnings
}
