package gps

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"gogator/internal/config"
)

func ReadRawCSV(path string, loc *time.Location, cfg config.Config) ([]RawPoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	first, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for i, h := range first {
		idx[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")))] = i
	}
	req := []string{"dt", "lat", "lng", "speed"}
	hasHeader := true
	for _, k := range req {
		if _, ok := idx[k]; !ok {
			hasHeader = false
			break
		}
	}
	if !hasHeader {
		idx = map[string]int{"dt": 0, "lat": 1, "lng": 2, "altitude": 3, "angle": 4, "speed": 5, "params": 6}
	}

	var out []RawPoint
	rawRow := 0
	if hasHeader {
		rawRow = 1
	}
	firstPending := !hasHeader
	for {
		var rec []string
		if firstPending {
			rec = first
			firstPending = false
		} else {
			var err error
			rec, err = r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
		}
		rawRow++
		get := func(k string) string {
			n, ok := idx[k]
			if !ok || n >= len(rec) {
				return ""
			}
			return rec[n]
		}
		dt := get("dt")
		lat, _ := strconv.ParseFloat(get("lat"), 64)
		lng, _ := strconv.ParseFloat(get("lng"), 64)
		alt, _ := strconv.ParseFloat(get("altitude"), 64)
		angle, _ := strconv.ParseFloat(get("angle"), 64)
		speed, _ := strconv.ParseFloat(get("speed"), 64)
		paramsRaw := get("params")
		params := ParseParams(paramsRaw)
		nums := NumericParams(params)
		t, err := parseTime(dt, loc, cfg)
		if err != nil {
			return nil, fmt.Errorf("row %d dt %q: %w", rawRow, dt, err)
		}
		out = append(out, RawPoint{RawRow: rawRow, RawDT: dt, Time: t, Lat: lat, Lng: lng, Altitude: alt, Angle: angle, SpeedKPH: speed, ParamsRaw: paramsRaw, Params: params, ParamNums: nums})
	}
	for i := range out {
		if i == 0 {
			continue
		}
		d := HaversineM(out[i-1].Lat, out[i-1].Lng, out[i].Lat, out[i].Lng)
		out[i].DistanceFromPrevM = d
		sec := out[i].Time.Sub(out[i-1].Time).Seconds()
		if sec > 0 {
			out[i].ImpliedSpeedKPH = d / sec * 3.6
		}
	}
	return out, nil
}

func parseTime(s string, loc *time.Location, cfg config.Config) (time.Time, error) {
	s = strings.TrimSpace(s)
	correction := time.Duration(cfg.RawTime.CorrectionHours * float64(time.Hour))

	if serial, err := strconv.ParseFloat(s, 64); err == nil {
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		dur := time.Duration(serial*24*float64(time.Hour)) + correction
		return base.Add(dur).In(loc), nil
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Add(correction).In(loc), nil
	}

	layouts := []string{"2006-01-02 15:04:05", "02/01/2006 15:04:05", "2/01/2006 15:04:05", "02/01/2006 3:04:05 PM"}
	for _, layout := range layouts {
		// Treat naive raw values as UTC tracker timestamps, then apply any
		// explicit legacy/emergency correction and render in the target timezone.
		// The default correction is 0; non-zero values are deliberately noisy
		// because they can move trips onto the wrong local calendar day.
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t.Add(correction).In(loc), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported datetime")
}
