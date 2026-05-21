package routes

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"gogator/internal/gps"
)

type Route struct {
	Name            string
	FromSite        string
	ToSite          string
	DistanceMinKM   float64
	DistanceMaxKM   float64
	DurationMinMin  float64
	DurationMaxMin  float64
	ConfidenceBoost string
	AutoMergeGapMin float64
	Notes           string
}

type Match struct {
	Name                  string
	Confidence            string
	Status                string
	ExpectedDistanceRange string
	ExpectedDurationRange string
	Notes                 string
	Matched               bool
}

type Observation struct {
	Index             int
	FromSite          string
	ToSite            string
	TripCount         int
	MedianDistanceKM  float64
	MinDistanceKM     float64
	MaxDistanceKM     float64
	MedianDurationMin float64
	MinDurationMin    float64
	MaxDurationMin    float64
	FirstSeen         time.Time
	LastSeen          time.Time
	SuggestedName     string
}

type Anomaly struct {
	TripIndex     int
	DepartureTime time.Time
	FromSite      string
	ToSite        string
	DistanceKM    float64
	DurationMin   float64
	RouteName     string
	Status        string
	Notes         string
	RawStartRow   int
	RawEndRow     int
}

func Load(path string) ([]Route, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	delimiter := detectDelimiter(data)
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	r.Comma = delimiter
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	headerIdx := -1
	for i, row := range rows {
		if hasHeader(row) {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return nil, nil
	}
	header := rows[headerIdx]
	idx := map[string]int{}
	for i, h := range header {
		idx[norm(h)] = i
	}
	get := func(row []string, keys ...string) string {
		for _, k := range keys {
			if n, ok := idx[norm(k)]; ok && n < len(row) {
				return strings.TrimSpace(row[n])
			}
		}
		return ""
	}
	var out []Route
	for _, row := range rows[headerIdx+1:] {
		from := get(row, "From Site", "From")
		to := get(row, "To Site", "To")
		if from == "" || to == "" {
			continue
		}
		name := get(row, "Route Name", "Name")
		if name == "" {
			name = from + " to " + to
		}
		out = append(out, Route{
			Name:            name,
			FromSite:        from,
			ToSite:          to,
			DistanceMinKM:   num(get(row, "Expected Distance Min Km", "Distance Min Km", "Min Distance Km")),
			DistanceMaxKM:   num(get(row, "Expected Distance Max Km", "Distance Max Km", "Max Distance Km")),
			DurationMinMin:  num(get(row, "Expected Duration Min Min", "Duration Min Min", "Min Duration Min")),
			DurationMaxMin:  num(get(row, "Expected Duration Max Min", "Duration Max Min", "Max Duration Min")),
			ConfidenceBoost: get(row, "Confidence Boost", "Confidence"),
			AutoMergeGapMin: num(get(row, "Auto Merge Gap Min", "Auto Merge Gap Minutes")),
			Notes:           get(row, "Notes"),
		})
	}
	return out, nil
}

func LoadObservations(path string) ([]Observation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	delimiter := detectDelimiter(data)
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	r.Comma = delimiter
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	headerIdx := -1
	for i, row := range rows {
		if hasObservationHeader(row) {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return nil, fmt.Errorf("no route observations header found in %s", path)
	}
	header := rows[headerIdx]
	idx := map[string]int{}
	for i, h := range header {
		idx[norm(h)] = i
	}
	get := func(row []string, keys ...string) string {
		for _, k := range keys {
			if n, ok := idx[norm(k)]; ok && n < len(row) {
				return strings.TrimSpace(row[n])
			}
		}
		return ""
	}
	var out []Observation
	fallbackIndex := 1
	for _, row := range rows[headerIdx+1:] {
		from := get(row, "From Site", "From")
		to := get(row, "To Site", "To")
		if from == "" || to == "" {
			continue
		}
		idxVal := int(num(get(row, "Index")))
		if idxVal <= 0 {
			idxVal = fallbackIndex
		}
		name := get(row, "Suggested Route Name", "Route Name", "Name")
		if name == "" {
			name = from + " to " + to
		}
		out = append(out, Observation{
			Index:             idxVal,
			FromSite:          from,
			ToSite:            to,
			TripCount:         int(num(get(row, "Trip Count"))),
			MedianDistanceKM:  num(get(row, "Median Distance Km")),
			MinDistanceKM:     num(get(row, "Min Distance Km")),
			MaxDistanceKM:     num(get(row, "Max Distance Km")),
			MedianDurationMin: num(get(row, "Median Duration Min")),
			MinDurationMin:    num(get(row, "Min Duration Min")),
			MaxDurationMin:    num(get(row, "Max Duration Min")),
			FirstSeen:         parseObservationTime(get(row, "First Seen")),
			LastSeen:          parseObservationTime(get(row, "Last Seen")),
			SuggestedName:     name,
		})
		fallbackIndex++
	}
	return out, nil
}

func FindObservation(observations []Observation, index int) (Observation, bool) {
	for _, o := range observations {
		if o.Index == index {
			return o, true
		}
	}
	return Observation{}, false
}

func RouteFromObservation(o Observation) Route {
	return Route{
		Name:            o.SuggestedName,
		FromSite:        o.FromSite,
		ToSite:          o.ToSite,
		DistanceMinKM:   round2(o.MinDistanceKM),
		DistanceMaxKM:   round2(o.MaxDistanceKM),
		DurationMinMin:  round2(o.MinDurationMin),
		DurationMaxMin:  round2(o.MaxDurationMin),
		ConfidenceBoost: "Observed",
		AutoMergeGapMin: 5,
		Notes:           fmt.Sprintf("Added from route observations index %d; trip_count=%d; median_distance_km=%.2f; median_duration_min=%.2f", o.Index, o.TripCount, o.MedianDistanceKM, o.MedianDurationMin),
	}
}

func AppendRoute(path string, route Route) error {
	delimiter := ','
	needsHeader := true
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		delimiter = detectDelimiter(data)
		needsHeader = !containsRouteHeader(data, delimiter)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if info, err := f.Stat(); err == nil && info.Size() > 0 {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	w := csv.NewWriter(f)
	w.Comma = delimiter
	if needsHeader {
		if err := w.Write(RouteHeaders()); err != nil {
			return err
		}
	}
	if err := w.Write(RouteRow(route)); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func RouteHeaders() []string {
	return []string{"Route Name", "From Site", "To Site", "Expected Distance Min Km", "Expected Distance Max Km", "Expected Duration Min Min", "Expected Duration Max Min", "Confidence Boost", "Auto Merge Gap Min", "Notes"}
}

func RouteRow(r Route) []string {
	return []string{r.Name, r.FromSite, r.ToSite, fmtFloat(r.DistanceMinKM), fmtFloat(r.DistanceMaxKM), fmtFloat(r.DurationMinMin), fmtFloat(r.DurationMaxMin), r.ConfidenceBoost, fmtFloat(r.AutoMergeGapMin), r.Notes}
}

func Apply(trips []gps.Trip, rules []Route, unknown string) ([]gps.Trip, []Observation, []Anomaly) {
	for i := range trips {
		m := MatchTrip(trips[i], rules, unknown)
		trips[i].RouteName = m.Name
		trips[i].RouteConfidence = m.Confidence
		trips[i].RouteMatchStatus = m.Status
		trips[i].RouteExpectedDistanceRange = m.ExpectedDistanceRange
		trips[i].RouteExpectedDurationRange = m.ExpectedDurationRange
		trips[i].RouteNotes = m.Notes
	}
	return trips, BuildObservations(trips, unknown), BuildAnomalies(trips, unknown)
}

func MatchTrip(t gps.Trip, rules []Route, unknown string) Match {
	from := strings.TrimSpace(t.DepartureSite)
	to := strings.TrimSpace(t.DestinationSite)
	if from == "" || to == "" || from == unknown || to == unknown {
		return Match{Confidence: "Review", Status: "Unknown endpoint", Notes: "Departure or destination site did not match a known site"}
	}
	for _, r := range rules {
		if same(r.FromSite, from) && same(r.ToSite, to) {
			m := Match{
				Name:                  r.Name,
				Confidence:            cleanConfidence(r.ConfidenceBoost, "Good"),
				Status:                "Matched expected route",
				ExpectedDistanceRange: rangeText(r.DistanceMinKM, r.DistanceMaxKM, "km"),
				ExpectedDurationRange: rangeText(r.DurationMinMin, r.DurationMaxMin, "min"),
				Notes:                 r.Notes,
				Matched:               true,
			}
			var problems []string
			if r.DistanceMinKM > 0 && t.DistanceKM < r.DistanceMinKM {
				problems = append(problems, "distance below expected")
			}
			if r.DistanceMaxKM > 0 && t.DistanceKM > r.DistanceMaxKM {
				problems = append(problems, "distance above expected")
			}
			durMin := t.DurationHours * 60
			if r.DurationMinMin > 0 && durMin < r.DurationMinMin {
				problems = append(problems, "duration below expected")
			}
			if r.DurationMaxMin > 0 && durMin > r.DurationMaxMin {
				problems = append(problems, "duration above expected")
			}
			if len(problems) > 0 {
				m.Confidence = "Review"
				m.Status = strings.Join(problems, "; ")
			}
			return m
		}
	}
	return Match{Confidence: "Unrated", Status: "No route rule", Notes: "Frequent routes are suggested in route_observations.csv"}
}

func BuildObservations(trips []gps.Trip, unknown string) []Observation {
	type bucket struct {
		from  string
		to    string
		dist  []float64
		dur   []float64
		first time.Time
		last  time.Time
	}
	buckets := map[string]*bucket{}
	for _, t := range trips {
		if !isObservationCandidate(t, unknown) {
			continue
		}
		key := t.DepartureSite + "\x00" + t.DestinationSite
		b := buckets[key]
		if b == nil {
			b = &bucket{from: t.DepartureSite, to: t.DestinationSite, first: t.Start, last: t.Start}
			buckets[key] = b
		}
		b.dist = append(b.dist, t.DistanceKM)
		b.dur = append(b.dur, t.DurationHours*60)
		if t.Start.Before(b.first) {
			b.first = t.Start
		}
		if t.Start.After(b.last) {
			b.last = t.Start
		}
	}
	var out []Observation
	for _, b := range buckets {
		out = append(out, Observation{
			FromSite:          b.from,
			ToSite:            b.to,
			TripCount:         len(b.dist),
			MedianDistanceKM:  median(b.dist),
			MinDistanceKM:     minv(b.dist),
			MaxDistanceKM:     maxv(b.dist),
			MedianDurationMin: median(b.dur),
			MinDurationMin:    minv(b.dur),
			MaxDurationMin:    maxv(b.dur),
			FirstSeen:         b.first,
			LastSeen:          b.last,
			SuggestedName:     b.from + " to " + b.to,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TripCount != out[j].TripCount {
			return out[i].TripCount > out[j].TripCount
		}
		if out[i].FromSite != out[j].FromSite {
			return out[i].FromSite < out[j].FromSite
		}
		return out[i].ToSite < out[j].ToSite
	})
	for i := range out {
		out[i].Index = i + 1
	}
	return out
}

func BuildAnomalies(trips []gps.Trip, unknown string) []Anomaly {
	var out []Anomaly
	for _, t := range trips {
		if !isRouteAnomaly(t, unknown) {
			continue
		}
		out = append(out, Anomaly{TripIndex: t.Index, DepartureTime: t.Start, FromSite: t.DepartureSite, ToSite: t.DestinationSite, DistanceKM: t.DistanceKM, DurationMin: t.DurationHours * 60, RouteName: t.RouteName, Status: t.RouteMatchStatus, Notes: t.RouteNotes, RawStartRow: t.RawStartRow, RawEndRow: t.RawEndRow})
	}
	return out
}

func isObservationCandidate(t gps.Trip, unknown string) bool {
	if t.DepartureSite == "" || t.DestinationSite == "" || t.DepartureSite == unknown || t.DestinationSite == unknown {
		return false
	}
	return !isRouteAnomaly(t, unknown)
}

func isRouteAnomaly(t gps.Trip, unknown string) bool {
	status := strings.TrimSpace(t.RouteMatchStatus)
	if t.DepartureSite == unknown || t.DestinationSite == unknown {
		return true
	}
	return status != "" && status != "Matched expected route" && status != "No route rule"
}

func detectDelimiter(data []byte) rune {
	for _, raw := range bytes.Split(data, []byte("\n")) {
		line := strings.TrimSpace(strings.TrimPrefix(string(raw), "\ufeff"))
		if line == "" {
			continue
		}
		tabs := strings.Count(line, "\t")
		commas := strings.Count(line, ",")
		semis := strings.Count(line, ";")
		if tabs >= commas && tabs >= semis && tabs > 0 {
			return '\t'
		}
		if semis > commas && semis > 0 {
			return ';'
		}
		return ','
	}
	return ','
}

func hasHeader(row []string) bool {
	hasFrom := false
	hasTo := false
	for _, c := range row {
		n := norm(c)
		if n == "from site" || n == "from" {
			hasFrom = true
		}
		if n == "to site" || n == "to" {
			hasTo = true
		}
	}
	return hasFrom && hasTo
}

func hasObservationHeader(row []string) bool {
	hasFrom := false
	hasTo := false
	hasTripCount := false
	for _, c := range row {
		n := norm(c)
		if n == "from site" || n == "from" {
			hasFrom = true
		}
		if n == "to site" || n == "to" {
			hasTo = true
		}
		if n == "trip count" {
			hasTripCount = true
		}
	}
	return hasFrom && hasTo && hasTripCount
}

func containsRouteHeader(data []byte, delimiter rune) bool {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	r.Comma = delimiter
	rows, err := r.ReadAll()
	if err != nil {
		return false
	}
	for _, row := range rows {
		if hasHeader(row) {
			return true
		}
	}
	return false
}

func norm(s string) string {
	s = strings.TrimPrefix(s, "\ufeff")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), " ")
}

func num(s string) float64 {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(s, "\ufeff"), "m"), "M"))
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func same(a, b string) bool { return norm(a) == norm(b) }

func rangeText(min, max float64, unit string) string {
	if min <= 0 && max <= 0 {
		return ""
	}
	if min > 0 && max > 0 {
		return fmt.Sprintf("%.2f-%.2f %s", min, max, unit)
	}
	if min > 0 {
		return fmt.Sprintf(">= %.2f %s", min, unit)
	}
	return fmt.Sprintf("<= %.2f %s", max, unit)
}

func cleanConfidence(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	a := append([]float64(nil), v...)
	sort.Float64s(a)
	mid := len(a) / 2
	if len(a)%2 == 1 {
		return a[mid]
	}
	return (a[mid-1] + a[mid]) / 2
}

func minv(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := math.Inf(1)
	for _, x := range v {
		if x < m {
			m = x
		}
	}
	return m
}

func maxv(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := math.Inf(-1)
	for _, x := range v {
		if x > m {
			m = x
		}
	}
	return m
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func fmtFloat(v float64) string {
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(round2(v), 'f', 2, 64)
}

func parseObservationTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
