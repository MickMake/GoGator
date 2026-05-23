package output

import (
	"encoding/csv"
	"os"
	"strings"
	"time"

	"gogator/engine"
)

type EngineDiagnosticPaths struct {
	Points, Motion, Stays, Visits, Excursions, CandidateTrips, TripComparison, ShadowSummary, ShadowMismatches string
}

type EngineDiagnosticOptions struct {
	Enabled, OutputDiagnostics, OutputPoints, OutputMotion, OutputStays, OutputVisits, OutputExcursions, OutputCandidateTrips, OutputTripComparison, OutputShadowSummary, OutputShadowMismatches bool
}

func WriteEngineDiagnostics(diag engine.Diagnostics, paths EngineDiagnosticPaths, o EngineDiagnosticOptions) error {
	if !o.Enabled {
		return nil
	}
	all := o.OutputDiagnostics
	if all || o.OutputPoints {
		if err := writePoints(paths.Points, diag.Points); err != nil {
			return err
		}
	}
	if all || o.OutputMotion {
		if err := writeMotion(paths.Motion, diag.Motion); err != nil {
			return err
		}
	}
	if all || o.OutputStays {
		if err := writeStays(paths.Stays, diag.Stays); err != nil {
			return err
		}
	}
	if all || o.OutputVisits {
		if err := writeVisits(paths.Visits, diag.Visits); err != nil {
			return err
		}
	}
	if all || o.OutputExcursions {
		if err := writeExcursions(paths.Excursions, diag.Excursions); err != nil {
			return err
		}
	}
	if all || o.OutputCandidateTrips {
		if err := writeCandidateTrips(paths.CandidateTrips, diag.CandidateTrips); err != nil {
			return err
		}
	}
	if all || o.OutputTripComparison {
		if err := writeTripComparison(paths.TripComparison, diag.TripComparison); err != nil {
			return err
		}
	}
	if all || o.OutputShadowSummary {
		if err := writeShadowSummary(paths.ShadowSummary, diag.ShadowSummary); err != nil {
			return err
		}
	}
	if all || o.OutputShadowMismatches {
		if err := writeShadowMismatches(paths.ShadowMismatches, diag.ShadowSummary.Mismatches); err != nil {
			return err
		}
	}
	return nil
}

func createCSV(path string, headers []string) (*csv.Writer, *os.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	w := csv.NewWriter(f)
	if err := w.Write(headers); err != nil {
		f.Close()
		return nil, nil, err
	}
	return w, f, nil
}
func flushClose(w *csv.Writer, f *os.File) error {
	w.Flush()
	if err := w.Error(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func writePoints(path string, pts []engine.PointEvidence) error {
	w, f, err := createCSV(path, []string{"source_index", "timestamp", "latitude", "longitude", "quality", "quality_score", "quality_reasons", "tracker_signals_summary"})
	if err != nil {
		return err
	}
	for _, p := range pts {
		reasons := make([]string, 0, len(p.Quality.Reasons))
		for _, r := range p.Quality.Reasons {
			reasons = append(reasons, string(r))
		}
		summary := []string{}
		if p.Signals.IO24 != nil {
			summary = append(summary, "io24")
		}
		if p.Signals.IO251 != nil {
			summary = append(summary, "io251")
		}
		if p.Signals.PDOP != nil {
			summary = append(summary, "pdop")
		}
		if err := w.Write([]string{itoa(p.Index), fmtTime(p.Time), ftoa(p.Coordinates.Lat, 6), ftoa(p.Coordinates.Lng, 6), string(p.Quality.Band), itoa(p.Quality.Score), strings.Join(reasons, ";"), strings.Join(summary, ";")}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
func writeMotion(path string, rows []engine.MotionSample) error {
	w, f, err := createCSV(path, []string{"source_index", "timestamp", "state", "speed_kmh", "reasons", "confidence"})
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{itoa(r.Index), fmtTime(r.Time), string(r.State), "", string(r.Reason), string(r.Quality)}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
func writeStays(path string, rows []engine.Stay) error {
	w, f, err := createCSV(path, []string{"stay_id", "start_time", "end_time", "duration_minutes", "latitude", "longitude", "radius_meters", "point_count", "stay_type", "confidence", "reasons", "matched_site"})
	if err != nil {
		return err
	}
	for i, r := range rows {
		reasons := make([]string, 0, len(r.Reasons))
		for _, x := range r.Reasons {
			reasons = append(reasons, string(x))
		}
		if err := w.Write([]string{itoa(i + 1), fmtTime(r.StartTime), fmtTime(r.EndTime), ftoa(r.Duration.Minutes(), 2), ftoa(r.Latitude, 6), ftoa(r.Longitude, 6), ftoa(r.RadiusM, 2), itoa(r.PointCount), string(r.Type), string(r.Confidence), strings.Join(reasons, ";"), r.MatchedSite}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
func writeVisits(path string, rows []engine.Visit) error {
	w, f, err := createCSV(path, []string{"visit_id", "stay_id", "start_time", "end_time", "duration_minutes", "visit_type", "matched_site", "confidence", "reasons"})
	if err != nil {
		return err
	}
	for i, r := range rows {
		reasons := []string{}
		for _, x := range r.Reasons {
			reasons = append(reasons, string(x))
		}
		if err := w.Write([]string{itoa(i + 1), itoa(r.StayIndex + 1), fmtTime(r.StartTime), fmtTime(r.EndTime), ftoa(r.Duration.Minutes(), 2), string(r.Type), r.MatchedSite, string(r.Confidence), strings.Join(reasons, ";")}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
func writeExcursions(path string, rows []engine.Excursion) error {
	w, f, err := createCSV(path, []string{"excursion_id", "from_visit_id", "to_visit_id", "start_time", "end_time", "duration_minutes", "distance_meters", "excursion_type", "confidence", "reasons"})
	if err != nil {
		return err
	}
	for i, r := range rows {
		reasons := []string{}
		for _, x := range r.Reasons {
			reasons = append(reasons, string(x))
		}
		if err := w.Write([]string{itoa(i + 1), itoa(r.FromVisitIndex + 1), itoa(r.ToVisitIndex + 1), fmtTime(r.StartTime), fmtTime(r.EndTime), ftoa(r.Duration.Minutes(), 2), ftoa(r.DistanceM, 2), string(r.Type), string(r.Confidence), strings.Join(reasons, ";")}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
func writeCandidateTrips(path string, rows []engine.CandidateTrip) error {
	w, f, err := createCSV(path, []string{"candidate_trip_id", "start_time", "end_time", "duration_minutes", "origin", "destination", "candidate_type", "confidence", "boundary_confidence", "reasons", "quality_warnings"})
	if err != nil {
		return err
	}
	for i, r := range rows {
		reasons := []string{}
		for _, x := range r.Reasons {
			reasons = append(reasons, string(x))
		}
		b := strings.Join([]string{string(r.OriginBoundary), string(r.DestinationBoundary), string(r.MovementBoundary), string(r.GPSQualityBoundary), string(r.GapNoiseBoundary), string(r.SiteMatchBoundary), string(r.DurationBoundary)}, "|")
		if err := w.Write([]string{itoa(i + 1), fmtTime(r.StartTime), fmtTime(r.EndTime), ftoa(r.Duration.Minutes(), 2), r.OriginLabel, r.DestinationLabel, string(r.Type), string(r.Confidence), b, strings.Join(reasons, ";"), strings.Join(r.Warnings, ";")}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
func writeTripComparison(path string, r engine.CandidateTripComparison) error {
	w, f, err := createCSV(path, []string{"metric", "value", "notes"})
	if err != nil {
		return err
	}
	rows := [][]string{{"candidate_trip_count", itoa(r.CandidateTripCount), ""}, {"legacy_valid_trip_count", itoa(r.LegacyValidTripCount), ""}, {"legacy_jitter_trip_count", itoa(r.LegacyJitterTripCount), ""}, {"approx_matched_trips", itoa(r.ApproxMatchedTrips), ""}}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}

func writeShadowSummary(path string, s engine.ShadowSummary) error {
	w, f, err := createCSV(path, []string{"metric", "value", "severity", "notes"})
	if err != nil {
		return err
	}
	for _, m := range s.Metrics {
		if err := w.Write([]string{m.Name, m.Value, string(m.Severity), m.Notes}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
func writeShadowMismatches(path string, mm []engine.ShadowMismatch) error {
	w, f, err := createCSV(path, []string{"mismatch_id", "mismatch_type", "severity", "legacy_start", "legacy_end", "candidate_start", "candidate_end", "delta_minutes", "notes"})
	if err != nil {
		return err
	}
	for i, m := range mm {
		id := m.ID
		if id == "" {
			id = itoa(i + 1)
		}
		if err := w.Write([]string{id, m.Type, string(m.Severity), fmtTime(m.LegacyStart), fmtTime(m.LegacyEnd), fmtTime(m.CandidateStart), fmtTime(m.CandidateEnd), ftoa(m.DeltaMinutes, 2), m.Notes}); err != nil {
			return err
		}
	}
	return flushClose(w, f)
}
