package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gogator/internal/gps"
	"gogator/internal/routes"
)

func Prefix(input string) string {
	dir := filepath.Dir(input)
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if dir == "." {
		return name
	}
	return filepath.Join(dir, name)
}

func WriteExpanded(path string, points []gps.RawPoint) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	headers := []string{"Filename", "Raw Row", "Raw Date/Time", "Normalised Date/Time", "Latitude", "Longitude", "Altitude", "Angle", "Speed kph", "Point Distance m", "Implied Speed kph", "Moving", "Stationary", "PDOP Quality", "Flags"}
	headers = append(headers, gps.ParamOrder...)
	headers = append(headers, "External Voltage V", "Backup Battery V", "Analog Input 1 V", "Accel Magnitude", "Params Raw")
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, p := range points {
		row := []string{p.SourceFile, itoa(p.RawRow), p.RawDT, formatTime(p.Time), ftoa(p.Lat, 6), ftoa(p.Lng, 6), ftoa(p.Altitude, 2), ftoa(p.Angle, 2), ftoa(p.SpeedKPH, 2), ftoa(p.DistanceFromPrevM, 2), ftoa(p.ImpliedSpeedKPH, 2), bools(p.Moving), bools(p.Stationary), p.PDOPQuality, strings.Join(p.Flags, ";")}
		for _, k := range gps.ParamOrder {
			row = append(row, p.Params[k])
		}
		row = append(row, div1000(p, "io66"), div1000(p, "io67"), div1000(p, "io6"))
		if m, ok := gps.AccelMagnitude(p.ParamNums); ok {
			row = append(row, ftoa(m, 2))
		} else {
			row = append(row, "")
		}
		row = append(row, p.ParamsRaw)
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return w.Error()
}

func WriteTrips(path string, trips []gps.Trip) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	headers := TripHeaders()
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, t := range trips {
		if err := w.Write(TripRow(t)); err != nil {
			return err
		}
	}
	return w.Error()
}

func TripHeaders() []string {
	h := []string{"Filename", "Import Index", "Job Number", "Departure Date/Time", "Departure Site", "Departure GPS", "Departure Address", "Travelling Duration", "Travelling Distance", "Travelling Top Speed", "Travelling Average Speed", "Destination Date/Time", "Destination Site", "Destination GPS", "Destination Address", "Site Duration", "Continuity Status", "Route Name", "Route Confidence", "Route Match Status", "Route Expected Distance Range", "Route Expected Duration Range", "Route Notes", "Raw Start Row", "Raw End Row", "Raw Points", "Flags"}
	extra := []string{"Ignition On Samples", "Ignition Off Samples", "Ignition Values", "Ignition Start", "Ignition End", "PDOP Poor Samples", "PDOP Ideal Samples", "GPS Level Max", "g0 Max Abs", "g1 Max Abs", "g2 Max Abs", "Accel Magnitude Max", "Accel Magnitude Avg", "Accel Nonzero Samples", "Crash Detected Samples", "Behaviour Event Samples", "Behaviour Event Values", "Panic Trigger Samples", "Power Cut Samples io252", "Sleep Mode Values", "Network State Values io381", "Trip Status Values io254", "Odometer Raw Start io14", "Odometer Raw End io14", "SIM ICCID Raw io11"}
	return append(h, extra...)
}

func TripRow(t gps.Trip) []string {
	sum := gps.TripParamSummary(t)
	status := t.ContinuityStatus
	if status == "" {
		status = "CONTINUITY_OK"
	}
	row := []string{
		t.Filename, itoa(t.Index), "", formatTime(t.Start), t.DepartureSite, gpsText(t.DepartLat, t.DepartLng), t.DepartureAddress,
		ftoa(t.DurationHours, 2), ftoa(t.DistanceKM, 2), ftoa(t.TopSpeedKPH, 2), ftoa(t.AverageSpeedKPH, 2),
		formatTime(t.End), t.DestinationSite, gpsText(t.DestLat, t.DestLng), t.DestinationAddress, ftoa(t.SiteDurationHours, 2), status,
		t.RouteName, t.RouteConfidence, t.RouteMatchStatus, t.RouteExpectedDistanceRange, t.RouteExpectedDurationRange, t.RouteNotes,
		itoa(t.RawStartRow), itoa(t.RawEndRow), itoa(t.RawPoints), strings.Join(t.Flags, ";"),
	}
	for _, k := range TripHeaders()[27:] {
		row = append(row, sum[k])
	}
	return row
}

func WriteRouteObservations(path string, observations []routes.Observation) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	headers := []string{"Index", "From Site", "To Site", "Trip Count", "Median Distance Km", "Min Distance Km", "Max Distance Km", "Median Duration Min", "Min Duration Min", "Max Duration Min", "First Seen", "Last Seen", "Suggested Route Name"}
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, o := range observations {
		row := []string{itoa(o.Index), o.FromSite, o.ToSite, itoa(o.TripCount), ftoa(o.MedianDistanceKM, 2), ftoa(o.MinDistanceKM, 2), ftoa(o.MaxDistanceKM, 2), ftoa(o.MedianDurationMin, 2), ftoa(o.MinDurationMin, 2), ftoa(o.MaxDurationMin, 2), formatTime(o.FirstSeen), formatTime(o.LastSeen), o.SuggestedName}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return w.Error()
}

func WriteRouteAnomalies(path string, anomalies []routes.Anomaly) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	headers := []string{"Trip Index", "Departure Date/Time", "From Site", "To Site", "Distance Km", "Duration Min", "Route Name", "Status", "Notes", "Raw Start Row", "Raw End Row"}
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, a := range anomalies {
		row := []string{itoa(a.TripIndex), formatTime(a.DepartureTime), a.FromSite, a.ToSite, ftoa(a.DistanceKM, 2), ftoa(a.DurationMin, 2), a.RouteName, a.Status, a.Notes, itoa(a.RawStartRow), itoa(a.RawEndRow)}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return w.Error()
}

func WriteAudit(path string, input, sites, routesPath, config, timezone string, rawCount, tripCount, jitterCount, siteCount, routeCount int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	rows := [][]string{{"Metric", "Value"}, {"Input", input}, {"Sites", sites}, {"Routes", routesPath}, {"Config", config}, {"Timezone", timezone}, {"Raw Points", itoa(rawCount)}, {"Valid Trips", itoa(tripCount)}, {"Rejected Jitter", itoa(jitterCount)}, {"Loaded Sites", itoa(siteCount)}, {"Loaded Routes", itoa(routeCount)}, {"Generated At", time.Now().Format(time.RFC3339)}}
	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return err
		}
	}
	return w.Error()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
func gpsText(lat, lng float64) string   { return fmt.Sprintf("%.6f,%.6f", lat, lng) }
func itoa(i int) string                 { return fmt.Sprintf("%d", i) }
func ftoa(f float64, places int) string { return fmt.Sprintf("%.*f", places, f) }
func bools(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}
func div1000(p gps.RawPoint, k string) string {
	if v, ok := gps.Num(p, k); ok {
		return ftoa(v/1000, 3)
	}
	return ""
}
